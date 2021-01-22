/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/suite_execution_volume"
	"github.com/palantir/stacktrace"
)

const (
	// Docker image that will be used to launch the container that will expand the files artifact
	//  into a Docker volume
	dockerImage = "alpine:3.12"

	// Dirpath on the artifact expander container where the suite execution volume (which contains the artifacts)
	//  will be mounted
	suiteExecutionVolumeMountDirpath = "/suite-execution"

	// Dirpath on the artifact expander container where the destination volume will be mounted
	destinationVolumeMountDirpath = "/dest"

	expanderContainerSuccessExitCode = 0
)

/*
Class responsible for taking an artifact containing compressed files and uncompressing its contents
	into a Docker volume that will be mounted on a new service
 */
type FilesArtifactExpander struct {
	suiteExecutionVolumeName string

	dockerManager *docker_manager.DockerManager

	testNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker
}

func NewFilesArtifactExpander(suiteExecutionVolumeName string, dockerManager *docker_manager.DockerManager, testNetworkId string, freeIpAddrTracker *commons.FreeIpAddrTracker) *FilesArtifactExpander {
	return &FilesArtifactExpander{suiteExecutionVolumeName: suiteExecutionVolumeName, dockerManager: dockerManager, testNetworkId: testNetworkId, freeIpAddrTracker: freeIpAddrTracker}
}


func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(ctx context.Context,
		artifactIdsToVolumeNames map[string]string) error {
	// Representation of the cache *on the expander image*
	expanderContainerArtifactCache := suite_execution_volume.NewArtifactCache(suiteExecutionVolumeMountDirpath)

	// TODO PERF: parallelize this to increase speed
	for artifactId, volumeName := range artifactIdsToVolumeNames {
		if err := expander.dockerManager.CreateVolume(ctx, volumeName); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the destination volume '%v'", volumeName)
		}

		artifactFilepathOnExpanderContainer := expanderContainerArtifactCache.GetArtifactFilepath(artifactId)

		containerCmd := getExtractionCommand(artifactFilepathOnExpanderContainer)

		volumeMounts := map[string]string{
			expander.suiteExecutionVolumeName: suiteExecutionVolumeMountDirpath,
			volumeName:                        destinationVolumeMountDirpath,
		}

		if err := expander.runExpanderContainer(ctx, containerCmd, volumeMounts); err != nil {
			return stacktrace.Propagate(err, "An error occurred running the expander container")
		}
	}

	return nil
}

// NOTE: This is a separate function so we can defer the releasing of the IP address and guarantee that it always
//  goes back into the IP pool
func (expander *FilesArtifactExpander) runExpanderContainer(
		ctx context.Context,
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
		expander.testNetworkId,
		containerIp,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		map[nat.Port]*nat.PortBinding{},
		containerCmd,
		map[string]string{}, // No env variables
		map[string]string{}, // No bind mounts
		volumeMounts,
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
		destinationVolumeMountDirpath,
	}
}



