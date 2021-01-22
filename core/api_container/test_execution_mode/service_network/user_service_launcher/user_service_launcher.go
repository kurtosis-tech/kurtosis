/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/volume_naming_consts"
	"github.com/palantir/stacktrace"
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

	freeIpAddrTracker *commons.FreeIpAddrTracker

	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander

	dockerNetworkId string

	// The name of the Docker volume containing data for this suite execution that will be mounted on this service
	suiteExecutionVolName string
}

func NewUserServiceLauncher(executionInstanceId string, testName string, dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander, dockerNetworkId string, suiteExecutionVolName string) *UserServiceLauncher {
	return &UserServiceLauncher{executionInstanceId: executionInstanceId, testName: testName, dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, filesArtifactExpander: filesArtifactExpander, dockerNetworkId: dockerNetworkId, suiteExecutionVolName: suiteExecutionVolName}
}


/**
Launches a testnet service with the given parameters

Returns: The container ID of the newly-launched service
 */
func (launcher UserServiceLauncher) Launch(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		ipAddr net.IP,
		imageName string,
		usedPorts map[nat.Port]bool,
		startCmd []string,
		dockerEnvVars map[string]string,
		suiteExecutionVolMntDirpath string,
		// Mapping artifactID -> mountpoint
		filesArtifactMountDirpaths map[string]string) (string, error) {

	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactIdsToVolumeNames := map[string]string{}
	for artifactId, _ := range filesArtifactMountDirpaths {
		artifactIdsToVolumeNames[artifactId] = launcher.getExpandedFilesArtifactVolName(
			serviceId,
			artifactId,
		)
	}
	if err := launcher.filesArtifactExpander.ExpandArtifactsIntoVolumes(ctx, artifactIdsToVolumeNames); err != nil {
		return "", stacktrace.Propagate(
			err,
			"An error occurred expanding the requested artifacts for service '%v' into Docker volumes",
			serviceId)
	}

	portBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		portBindings[port] = nil
	}

	volumeMounts := map[string]string{
		launcher.suiteExecutionVolName: suiteExecutionVolMntDirpath,
	}
	for artifactId, mountpoint := range filesArtifactMountDirpaths {
		volumeName, found := artifactIdsToVolumeNames[artifactId]
		if !found {
			return "", stacktrace.NewError(
				"Could not find a volume name corresponding to artifact ID '%v' AFTER expansion; this is very strange",
				artifactId)
		}
		volumeMounts[volumeName] = mountpoint
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		ctx,
		imageName,
		launcher.dockerNetworkId,
		ipAddr,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		portBindings,
		startCmd,
		dockerEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		volumeMounts,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", imageName)
	}
	return containerId, nil
}

// ==================================================================================================
//                                     Private helper functions
// ==================================================================================================
func (launcher UserServiceLauncher) getExpandedFilesArtifactVolName(
		serviceId service_network_types.ServiceID,
		artifactId string) string {
	timestampStr := time.Now().Format(volume_naming_consts.GoTimestampFormat)
	return fmt.Sprintf(
		"%v_%v_%v_%v_%v",
		timestampStr,
		launcher.executionInstanceId,
		launcher.testName,
		serviceId,
		artifactId)
}
