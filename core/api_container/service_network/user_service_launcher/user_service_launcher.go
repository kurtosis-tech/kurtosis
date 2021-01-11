/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/topology_types"
	"github.com/kurtosis-tech/kurtosis/api_container/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/volume_naming_consts"
	"github.com/palantir/stacktrace"
	"net"
	"strings"
	"time"
)

const (
	// We use YYYY-MM-DDTHH.mm.ss to match the timestamp format that the suite execution volume uses
	expandedFilesArtifactVolNameDatePattern = "2006-01-02T15.04.05"
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
	testVolumeName string
}

func NewUserServiceLauncher(executionInstanceId string, testName string, dockerManager *docker_manager.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander, dockerNetworkId string, testVolumeName string) *UserServiceLauncher {
	return &UserServiceLauncher{executionInstanceId: executionInstanceId, testName: testName, dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, filesArtifactExpander: filesArtifactExpander, dockerNetworkId: dockerNetworkId, testVolumeName: testVolumeName}
}


/**
Launches a testnet service with the given parameters

Returns: The container ID of the newly-launched service
 */
func (launcher UserServiceLauncher) Launch(
		ctx context.Context,
		serviceId topology_types.ServiceID,
		ipAddr net.IP,
		imageName string,
		usedPorts map[nat.Port]bool,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountDirpath string,
		// Mapping artifactID -> mountpoint
		filesArtifactMountDirpaths map[string]string) (string, error) {

	// TODO keep track of the volumes we create, and delete them at the end of the test to keep things cleaner
	// First expand the files artifacts into volumes, so that any errors get caught early
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

	// The user won't know the IP address, so we'll need to replace all the IP address placeholders with the actual
	//  IP
	replacedStartCmd, replacedEnvVars := replaceIpPlaceholderForDockerParams(
		ipPlaceholder,
		ipAddr,
		startCmd,
		dockerEnvVars)

	portBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		portBindings[port] = nil
	}

	volumeMounts := map[string]string{
		launcher.testVolumeName: testVolumeMountDirpath,
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
		replacedStartCmd,
		replacedEnvVars,
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
		serviceId topology_types.ServiceID,
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

/*
Small helper function to replace the IP placeholder with the real IP string in the start command and Docker environment
	variables.
*/
func replaceIpPlaceholderForDockerParams(
		ipPlaceholder string,
		realIp net.IP,
		startCmd []string,
		envVars map[string]string) ([]string, map[string]string) {
	ipPlaceholderStr := ipPlaceholder
	replacedStartCmd := []string{}
	for _, cmdFragment := range startCmd {
		replacedCmdFragment := strings.ReplaceAll(cmdFragment, ipPlaceholderStr, realIp.String())
		replacedStartCmd = append(replacedStartCmd, replacedCmdFragment)
	}
	replacedEnvVars := map[string]string{}
	for key, value := range envVars {
		replacedValue := strings.ReplaceAll(value, ipPlaceholderStr, realIp.String())
		replacedEnvVars[key] = replacedValue
	}
	return replacedStartCmd, replacedEnvVars
}
