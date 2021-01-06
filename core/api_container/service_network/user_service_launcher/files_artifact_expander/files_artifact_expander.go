/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/artifact_cache"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/files_artifact_expander_consts"
	"github.com/palantir/stacktrace"
)

const (
	successExitCode = 0
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
	expanderContainerArtifactCache := artifact_cache.NewArtifactCache(files_artifact_expander_consts.SuiteExecutionVolumeMountDirpath)

	// TODO parallelize this to increase speed
	for artifactId, volumeName := range artifactIdsToVolumeNames {
		if err := expander.dockerManager.CreateVolume(ctx, volumeName); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating the destination volume '%v'", volumeName)
		}

		artifactFilepathOnExpanderContainer := expanderContainerArtifactCache.GetArtifactFilepath(artifactId)

		containerCmd := files_artifact_expander_consts.GetExtractionCommand(artifactFilepathOnExpanderContainer)

		volumeMounts := map[string]string{
			expander.suiteExecutionVolumeName: files_artifact_expander_consts.SuiteExecutionVolumeMountDirpath,
			volumeName:                        files_artifact_expander_consts.DestinationVolumeMountDirpath,
		}

		// TODO This silently (temporarily) uses up one of the user's requested IP addresses with a container
		//  that's not one of their services! This could get confusing if the user requests exactly a wide enough
		//  subnet to fit all _their_ services, but we hit the limit because we have these admin containers too
		//  Possible fix: create a special "admin" network, one per suite execution, for doing thinks like this?
		containerIp, err := expander.freeIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting a free IP for the file artifact-expanding container")
		}

		containerId, err := expander.dockerManager.CreateAndStartContainer(
			ctx,
			files_artifact_expander_consts.DockerImage,
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
		if exitCode != successExitCode {
			return stacktrace.NewError(
				"The files artifact-expanding Docker container exited with non-%v exit code: ",
				successExitCode,
				exitCode)
		}

		// TODO release the IP we acquired from the free IP addr tracker
	}

	return nil
}



