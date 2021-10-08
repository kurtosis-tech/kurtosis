/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/commons"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_data_volume"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
	"github.com/palantir/stacktrace"
	"path"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data volume (which contains the artifacts)
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
	enclaveDataVolName string

	dockerManager *docker_manager.DockerManager

	enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider

	testNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	filesArtifactCache *enclave_data_volume.FilesArtifactCache
}

func NewFilesArtifactExpander(enclaveDataVolName string, dockerManager *docker_manager.DockerManager, enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider, testNetworkId string, freeIpAddrTracker *commons.FreeIpAddrTracker, filesArtifactCache *enclave_data_volume.FilesArtifactCache) *FilesArtifactExpander {
	return &FilesArtifactExpander{enclaveDataVolName: enclaveDataVolName, dockerManager: dockerManager, enclaveObjNameProvider: enclaveObjNameProvider, testNetworkId: testNetworkId, freeIpAddrTracker: freeIpAddrTracker, filesArtifactCache: filesArtifactCache}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
		ctx context.Context,
		serviceGUID service_network_types.ServiceGUID, // Service GUID for whom the artifacts are being expanded into volumes
		artifactIdsToExpand map[string]bool) (map[string]string, error) {
	artifactIdsToVolNames := map[string]string{}
	for artifactId := range artifactIdsToExpand {
		destVolName := expander.enclaveObjNameProvider.ForFilesArtifactExpansionVolume(serviceGUID, artifactId)
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

		containerName := expander.enclaveObjNameProvider.ForFilesArtifactExpanderContainer(serviceGUID, artifactId)
		if err := expander.runExpanderContainer(ctx, containerName, containerCmd, volumeMounts); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running the expander container")
		}
	}

	return artifactIdsToVolNames, nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
// NOTE: This is a separate function so we can defer the releasing of the IP address and guarantee that it always
//  goes back into the IP pool
func (expander *FilesArtifactExpander) runExpanderContainer(
		ctx context.Context,
		containerName string,
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

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		dockerImage,
		containerName,
		expander.testNetworkId,
	).WithStaticIP(
		containerIp,
	).WithCmdArgs(
		containerCmd,
	).WithVolumeMounts(
		volumeMounts,
	).Build()
	containerId, _, err := expander.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
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



