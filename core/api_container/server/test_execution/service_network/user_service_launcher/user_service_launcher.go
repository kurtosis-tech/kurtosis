/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/free_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/kurtosis-tech/kurtosis/commons/volume_naming_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

/*
Convenience struct whose only purpose is launching user services
 */
type UserServiceLauncher struct {
	executionInstanceId string

	testName string

	dockerManager *docker_manager.DockerManager

	containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker

	// A nil value for this field indicates that no service port <-> host port bindings should be done
	freeHostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier

	artifactCache *suite_execution_volume.ArtifactCache

	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander

	dockerNetworkId string

	// The name of the Docker volume containing data for this suite execution that will be mounted on this service
	suiteExecutionVolName string
}

func NewUserServiceLauncher(executionInstanceId string, testName string, dockerManager *docker_manager.DockerManager, containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider, freeIpAddrTracker *commons.FreeIpAddrTracker, freeHostPortBindingSupplier *free_host_port_binding_supplier.FreeHostPortBindingSupplier, artifactCache *suite_execution_volume.ArtifactCache, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander, dockerNetworkId string, suiteExecutionVolName string) *UserServiceLauncher {
	return &UserServiceLauncher{executionInstanceId: executionInstanceId, testName: testName, dockerManager: dockerManager, containerNameElemsProvider: containerNameElemsProvider, freeIpAddrTracker: freeIpAddrTracker, freeHostPortBindingSupplier: freeHostPortBindingSupplier, artifactCache: artifactCache, filesArtifactExpander: filesArtifactExpander, dockerNetworkId: dockerNetworkId, suiteExecutionVolName: suiteExecutionVolName}
}

/**
Launches a testnet service with the given parameters

Returns:
	* The container ID of the newly-launched service
	* The mapping of used_port -> host_port_binding (if any)
 */
func (launcher UserServiceLauncher) Launch(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		ipAddr net.IP,
		imageName string,
		usedPorts map[nat.Port]bool,
		entrypointArgs []string,
		cmdArgs []string,
		dockerEnvVars map[string]string,
		suiteExecutionVolMntDirpath string,
		// Mapping artifactUrl -> mountpoint
		artifactUrlToMountDirpath map[string]string) (string, map[nat.Port]*nat.PortBinding, error) {
	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactToVolName := map[suite_execution_volume.Artifact]string{}
	artifactVolToMountpoint := map[string]string{}
	for artifactUrl, mountDirpath := range artifactUrlToMountDirpath {
		logrus.Debugf("Hashing artifact URL '%v' to be mounted at '%v'...", artifactUrl, mountDirpath)
		artifact, err := launcher.artifactCache.GetArtifact(artifactUrl)
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "An error occurred getting artifact with URL '%v' from artifact cache", artifactUrl)
		}
		artifactUrlHash := artifact.GetUrlHash()
		destVolName := launcher.getExpandedFilesArtifactVolName(
			serviceId,
			artifactUrlHash)
		artifactToVolName[*artifact] = destVolName
		artifactVolToMountpoint[destVolName] = mountDirpath
	}
	if err := launcher.filesArtifactExpander.ExpandArtifactsIntoVolumes(ctx, serviceId, artifactToVolName); err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"An error occurred expanding the requested artifacts for service '%v' into Docker volumes",
			serviceId)
	}

	hostPortBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		if launcher.freeHostPortBindingSupplier != nil {
			binding, err := launcher.freeHostPortBindingSupplier.GetFreePortBinding()
			if err != nil {
				return "", nil, stacktrace.Propagate(
					err,
					"Host port binding was requested, but an error occurred getting a free host port to bind to service port %v",
					port.Port(),
				)
			}
			hostPortBindings[port] = &binding
		}
	}

	volumeMounts := map[string]string{
		launcher.suiteExecutionVolName: suiteExecutionVolMntDirpath,
	}
	for artifactVolName, mountpoint := range artifactVolToMountpoint {
		volumeMounts[artifactVolName] = mountpoint
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		ctx,
		imageName,
		launcher.containerNameElemsProvider.GetForUserService(serviceId),
		launcher.dockerNetworkId,
		ipAddr,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		hostPortBindings,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		volumeMounts,
		false,		// User services definitely shouldn't be able to access the Docker host machine
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", imageName)
	}
	return containerId, hostPortBindings, nil
}

// ==================================================================================================
//                                     Private helper functions
// ==================================================================================================
func (launcher UserServiceLauncher) getExpandedFilesArtifactVolName(
		serviceId service_network_types.ServiceID,
		artifactUrlHash string) string {
	timestampStr := time.Now().Format(volume_naming_consts.GoTimestampFormat)
	return fmt.Sprintf(
		"%v_%v_%v_%v_%v",
		timestampStr,
		launcher.executionInstanceId,
		launcher.testName,
		serviceId,
		artifactUrlHash)
}
