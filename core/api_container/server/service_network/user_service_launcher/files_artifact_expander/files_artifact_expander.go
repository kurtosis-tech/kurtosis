/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/container_name_provider"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_data_volume"
	"github.com/kurtosis-tech/kurtosis/commons/volume_naming_consts"
	"github.com/palantir/stacktrace"
	"path"
	"strings"
	"time"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the suite execution volume (which contains the artifacts)
	//  will be mounted
	enclaveDataVolMountpointOnExpanderContainer = "/enclave-data"

	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	expanderContainerSuccessExitCode = 0
)

/*
Class responsible for taking an artifact containing compressed files and uncompressing its contents
	into a Docker volume that will be mounted on a new service
 */
type FilesArtifactExpander struct {
	enclaveNameElems []string

	enclaveDataVolName string

	dockerManager *docker_manager.DockerManager

	containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider

	testNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	filesArtifactCache *enclave_data_volume.FilesArtifactCache
}

func NewFilesArtifactExpander(enclaveNameElems []string, enclaveDataVolName string, dockerManager *docker_manager.DockerManager, containerNameElemsProvider *container_name_provider.ContainerNameElementsProvider, testNetworkId string, freeIpAddrTracker *commons.FreeIpAddrTracker, filesArtifactCache *enclave_data_volume.FilesArtifactCache) *FilesArtifactExpander {
	return &FilesArtifactExpander{enclaveNameElems: enclaveNameElems, enclaveDataVolName: enclaveDataVolName, dockerManager: dockerManager, containerNameElemsProvider: containerNameElemsProvider, testNetworkId: testNetworkId, freeIpAddrTracker: freeIpAddrTracker, filesArtifactCache: filesArtifactCache}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
		ctx context.Context,
		serviceId service_network_types.ServiceID, // Service ID for whom the artifacts are being expanded into volumes
		artifactIdsToExpand map[string]bool) (map[string]string, error) {
	artifactIdsToVolNames := map[string]string{}
	for artifactId := range artifactIdsToExpand {
		destVolName := expander.getExpandedFilesArtifactVolName(serviceId, artifactId)
		artifactIdsToVolNames[artifactId] = destVolName
	}

	// TODO PERF: parallelize this to increase speed
	for artifactId, destVolName := range artifactIdsToVolNames {
		artifactFile, err := expander.filesArtifactCache.GetFilesArtifact(artifactId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", artifactId)
		}

		if err := expander.dockerManager.CreateVolume(ctx, destVolName); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the destination volume '%v'", destVolName)
		}

		artifactRelativeFilepath := artifactFile.GetFilepathRelativeToVolRoot()
		artifactFilepathOnExpanderContainer := path.Join(
			enclaveDataVolMountpointOnExpanderContainer,
			artifactRelativeFilepath,
		)

		containerCmd := getExtractionCommand(artifactFilepathOnExpanderContainer)

		volumeMounts := map[string]string{
			expander.enclaveDataVolName: enclaveDataVolMountpointOnExpanderContainer,
			destVolName:                  destVolMntDirpathOnExpander,
		}

		containerNameElems := expander.containerNameElemsProvider.GetForFilesArtifactExpander(serviceId, artifactId)
		if err := expander.runExpanderContainer(ctx, containerNameElems, containerCmd, volumeMounts); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running the expander container")
		}
	}

	return artifactIdsToVolNames, nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func (expander *FilesArtifactExpander) getExpandedFilesArtifactVolName(
	serviceId service_network_types.ServiceID,
	artifactId string) string {
	timestampStr := time.Now().Format(volume_naming_consts.GoTimestampFormat)
	prefix := strings.Join(expander.enclaveNameElems, "_")
	// TODO Standardize this!
	return fmt.Sprintf(
		"%v_%v_%v_%v",
		timestampStr,
		prefix,
		serviceId,
		artifactId,
	)
}

// NOTE: This is a separate function so we can defer the releasing of the IP address and guarantee that it always
//  goes back into the IP pool
func (expander *FilesArtifactExpander) runExpanderContainer(
		ctx context.Context,
		containerNameElems []string,
		containerCmd []string,
		volumeMounts map[string]string) error {
	// NOTE: This silently (temporarily) uses up one of the user's requested IP addresses with a container
	//  that's not one of their services! This could get confusing if the user requests exactly a wide enough
	//  subnet to fit all _their_ services, but we hit the limit because we have these admin containers too
	//  If this becomes problematic, create a special "admin" network, one per suite execution, for doing thinks like this?
	containerIp, err := expander.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a free IP for the file artifact-expanding container")
	}
	defer expander.freeIpAddrTracker.ReleaseIpAddr(containerIp)

	containerId, err := expander.dockerManager.CreateAndStartContainer(
		ctx,
		dockerImage,
		containerNameElems,
		expander.testNetworkId,
		containerIp,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{},
		nil, // No ENTRYPOINT overriding needed
		containerCmd,
		map[string]string{}, // No env variables
		map[string]string{}, // No bind mounts
		volumeMounts,
		false,		// Files artifact expander doesn't need access to the Docker host machine
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker container to expand the artifact into the volume")
	}

	exitCode, err := expander.dockerManager.WaitForExit(ctx, containerId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the files artifact-expanding Docker container to exit")
	}
	if exitCode != expanderContainerSuccessExitCode {
		return stacktrace.NewError(
			"The files artifact-expanding Docker container exited with non-%v exit code: %v",
			expanderContainerSuccessExitCode,
			exitCode)
	}
	return nil
}

// Image-specific generator of the command that should be run to extract the artifact at the given filepath
//  to the destination
func getExtractionCommand(artifactFilepath string) (dockerRunCmd []string) {
	return []string{
		"tar",
		"-xzvf",
		artifactFilepath,
		"-C",
		destVolMntDirpathOnExpander,
	}
}



