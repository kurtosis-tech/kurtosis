/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	// This should probably be configurable
	shouldPullImageBeforeLaunch = true

	// User services shouldn't have access to the Docker engine
	shouldBindMountDockerSocket = false
)

// 1:1 mapping between enclave container port protos <-> obj attrs schema protos
var enclaveContainerPortProtosToObjAttrsPortProtos = map[enclave_container_launcher.EnclaveContainerPortProtocol]schema.PortProtocol{
	enclave_container_launcher.EnclaveContainerPortProtocol_TCP: schema.PortProtocol_TCP,
	enclave_container_launcher.EnclaveContainerPortProtocol_SCTP: schema.PortProtocol_SCTP,
	enclave_container_launcher.EnclaveContainerPortProtocol_UDP: schema.PortProtcol_UDP,
}

/*
Convenience struct whose only purpose is launching user services
 */
type UserServiceLauncher struct {
	dockerManager *docker_manager.DockerManager
	
	enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher

	freeIpAddrTracker *lib.FreeIpAddrTracker

	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander
}

func NewUserServiceLauncher(dockerManager *docker_manager.DockerManager, enclaveContainerLauncher *enclave_container_launcher.EnclaveContainerLauncher, freeIpAddrTracker *lib.FreeIpAddrTracker, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander) *UserServiceLauncher {
	return &UserServiceLauncher{dockerManager: dockerManager, enclaveContainerLauncher: enclaveContainerLauncher, freeIpAddrTracker: freeIpAddrTracker, filesArtifactExpander: filesArtifactExpander}
}

/**
Launches a testnet service with the given parameters

Returns:
	* The container ID of the newly-launched service
	* The mapping of used_port -> host_port_binding (if no host port is bound, then the value will be nil)
 */
func (launcher UserServiceLauncher) Launch(
	ctx context.Context,
	serviceGUID service_network_types.ServiceGUID,
	dockerContainerAlias string,
	ipAddr net.IP,
	imageName string,
	dockerNetworkId string,
	privatePorts map[string]*enclave_container_launcher.EnclaveContainerPort,
	entrypointArgs []string,
	cmdArgs []string,
	dockerEnvVars map[string]string,
	enclaveDataDirMountDirpath string,
	// Mapping files artifact ID -> mountpoint on the container to launch
	filesArtifactIdsToMountpoints map[string]string,
) (
	resultContainerId string,
	resultPublicIpAddr net.IP,	// Will be nil if len(privatePorts) == 0
	resultPublicPorts map[string]*enclave_container_launcher.EnclaveContainerPort, // Will be empty if len(privatePorts) == 0
	resultErr error,
) {
	allObjAttrsPorts := map[string]*schema.PortSpec{}
	for portId, enclaveContainerPort := range privatePorts {
		portNum := enclaveContainerPort.GetNumber()
		enclaveContainerPortProto := enclaveContainerPort.GetProtocol()
		objAttrsPortProto, found := enclaveContainerPortProtosToObjAttrsPortProtos[enclaveContainerPortProto]
		if !found {
			return "", nil, nil, stacktrace.NewError(
				"No object attributes schema port protocol found for enclave container port protocol '%v' used by port '%v'; this is a bug in Kurtosis",
				enclaveContainerPortProto,
				portId,
			)
		}
		objAttrsPort, err := schema.NewPortSpec(
			portNum,
			objAttrsPortProto,
		)
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(
				err,
				"An error occurred constructing object attributes port spec using port num '%v' and protocol '%v'",
				portNum,
				objAttrsPortProto,
			)
		}
		allObjAttrsPorts[portId] = objAttrsPort
	}

	//We use serviceID as the container alias
	serviceId := dockerContainerAlias
	objAttrsSupplier := func(enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error) {
		userServiceContainerAttrs, err := enclaveObjAttrsProvider.ForUserServiceContainer(
			serviceId,
			string(serviceGUID),
			allObjAttrsPorts,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred getting the user service container object attributes using service ID '%v' and private ports '%+v'",
				serviceGUID,
				allObjAttrsPorts,
			)
		}
		return userServiceContainerAttrs, nil
	}

	usedArtifactIdSet := map[files_artifact.FilesArtifactID]bool{}
	for artifactIdStr := range filesArtifactIdsToMountpoints {
		artifactId := files_artifact.FilesArtifactID(artifactIdStr)
		usedArtifactIdSet[artifactId] = true
	}

	//TODO we should remove this var when whe replace `service_network_types.ServiceGUID` with `service.ServiceGUID`
	//TODO in all this project
	adaptedServiceGuid := service.ServiceGUID(serviceGUID)
	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactIdsToVolumes, err := launcher.filesArtifactExpander.ExpandArtifactsIntoVolumes(ctx, adaptedServiceGuid, usedArtifactIdSet)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred expanding the requested files artifacts into volumes")
	}

	artifactVolumeMounts := map[string]string{}
	for artifactIdStr, mountpoint := range filesArtifactIdsToMountpoints {
		artifactId := files_artifact.FilesArtifactID(artifactIdStr)
		artifactVolumeName, found := artifactIdsToVolumes[artifactId]
		if !found {
			return "", nil, nil, stacktrace.NewError(
				"Even though we declared that we need files artifact '%v' to be expanded, no volume containing the " +
					"expanded contents was found; this is a bug in Kurtosis",
				artifactId,
			)
		}
		artifactVolumeNameStr := string(artifactVolumeName)
		artifactVolumeMounts[artifactVolumeNameStr] = mountpoint
	}

	containerId, publicIpAddr, publicPorts, err := launcher.enclaveContainerLauncher.Launch(
		ctx,
		imageName,
		shouldPullImageBeforeLaunch,
		ipAddr,
		dockerNetworkId,
		enclaveDataDirMountDirpath,
		privatePorts,
		objAttrsSupplier,
		dockerEnvVars,
		shouldBindMountDockerSocket,
		dockerContainerAlias,
		entrypointArgs,
		cmdArgs,
		artifactVolumeMounts,
	)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for user service with image '%v'", imageName)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Error("Launching the service container failed, but an error occurred killing container we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill container with ID '%v'", containerId)
			}
		}
	}()

	shouldKillContainer = false
	return containerId, publicIpAddr, publicPorts, nil
}
