/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the enclave data dir (which contains the artifacts)
	//  will be bind-mounted
	enclaveDataDirMountpointOnExpanderContainer = "/enclave-data"

	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	expanderContainerSuccessExitCode = 0
)

/*
Class responsible for taking an artifact containing compressed files and uncompressing its contents
	into a Docker volume that will be mounted on a new service
 */
type FilesArtifactExpander struct {
	// Host machine dirpath so the expander can bind-mount it to the artifact expansion containers
	enclaveDataDirpathOnHostMachine string

	dockerManager *docker_manager.DockerManager

	enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider

	testNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	filesArtifactCache *enclave_data_directory.FilesArtifactCache
}

func NewFilesArtifactExpander(enclaveDataDirpathOnHostMachine string, dockerManager *docker_manager.DockerManager, enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider, testNetworkId string, freeIpAddrTracker *commons.FreeIpAddrTracker, filesArtifactCache *enclave_data_directory.FilesArtifactCache) *FilesArtifactExpander {
	return &FilesArtifactExpander{enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine, dockerManager: dockerManager, enclaveObjAttrsProvider: enclaveObjAttrsProvider, testNetworkId: testNetworkId, freeIpAddrTracker: freeIpAddrTracker, filesArtifactCache: filesArtifactCache}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
		ctx context.Context,
		serviceGUID service_network_types.ServiceGUID, // Service GUID for whom the artifacts are being expanded into volumes
		artifactIdsToExpand map[string]bool,
) (map[string]string, error) {
	artifactIdsToVolAttrs := map[string]schema.ObjectAttributes{}
	for artifactId := range artifactIdsToExpand {
		destVolAttrs := expander.enclaveObjAttrsProvider.ForFilesArtifactExpansionVolume(string(serviceGUID), artifactId)
		artifactIdsToVolAttrs[artifactId] = destVolAttrs
	}

	// TODO PERF: parallelize this to increase speed
	artifactIdsToVolNames := map[string]string{}
	for artifactId, destVolAttrs := range artifactIdsToVolAttrs {
		artifactFile, err := expander.filesArtifactCache.GetFilesArtifact(artifactId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", artifactId)
		}

		volumeName := destVolAttrs.GetName()
		volumeLabels := destVolAttrs.GetLabels()
		if err := expander.dockerManager.CreateVolume(ctx, volumeName, volumeLabels); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the destination volume '%v' with labels '%+v'", volumeName, volumeLabels)
		}

		artifactRelativeFilepath := artifactFile.GetFilepathRelativeToDataDirRoot()
		artifactFilepathOnExpanderContainer := path.Join(
			enclaveDataDirMountpointOnExpanderContainer,
			artifactRelativeFilepath,
		)

		containerCmd := getExtractionCommand(artifactFilepathOnExpanderContainer)

		volumeMounts := map[string]string{
			volumeName: destVolMntDirpathOnExpander,
		}
		containerAttrs := expander.enclaveObjAttrsProvider.ForFilesArtifactExpanderContainer(string(serviceGUID), artifactId)
		containerName := containerAttrs.GetName()
		containerLabels := containerAttrs.GetLabels()
		if err := expander.runExpanderContainer(ctx, containerName, containerCmd, volumeMounts, containerLabels); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running the expander container")
		}

		artifactIdsToVolNames[artifactId] = volumeName
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
		volumeMounts map[string]string,
		labels map[string]string,
) error {
	// NOTE: This silently (temporarily) uses up one of the user's requested IP addresses with a container
	//  that's not one of their services! This could get confusing if the user requests exactly a wide enough
	//  subnet to fit all _their_ services, but we hit the limit because we have these admin containers too
	//  If this becomes problematic, create a special "admin" network, one per suite execution, for doing thinks like this?
	containerIp, err := expander.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a free IP for the file artifact-expanding container")
	}
	defer expander.freeIpAddrTracker.ReleaseIpAddr(containerIp)

	bindMounts := map[string]string{
		expander.enclaveDataDirpathOnHostMachine: enclaveDataDirMountpointOnExpanderContainer,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		dockerImage,
		containerName,
		expander.testNetworkId,
	).WithStaticIP(
		containerIp,
	).WithCmdArgs(
		containerCmd,
	).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		labels,
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



