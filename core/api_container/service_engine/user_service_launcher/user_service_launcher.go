/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
)

/*
Convenience struct whose only purpose is launching user services
 */
type UserServiceLauncher struct {
	dockerManager *commons.DockerManager

	freeIpAddrTracker *commons.FreeIpAddrTracker

	dockerNetworkId string

	// The name of the Docker volume for this test that will be mounted on all testnet services
	testVolumeName string
}

func NewUserServiceLauncher(dockerManager *commons.DockerManager, freeIpAddrTracker *commons.FreeIpAddrTracker, dockerNetworkId string, testVolumeName string) *UserServiceLauncher {
	return &UserServiceLauncher{dockerManager: dockerManager, freeIpAddrTracker: freeIpAddrTracker, dockerNetworkId: dockerNetworkId, testVolumeName: testVolumeName}
}

/**
Launches a testnet service with the given parameters

Returns: The container ID of the newly-launched service, and the IP that the service was launched with
 */
func (launcher UserServiceLauncher) Launch(
		context context.Context,
		imageName string,
		usedPorts map[nat.Port]bool,
		ipPlaceholder string,
		startCmd []string,
		dockerEnvVars map[string]string,
		testVolumeMountFilepath string) (string, net.IP, error) {
	freeIp, err := launcher.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"An error occurred when getting an IP to give the container running the new service with Docker image '%v'",
			imageName)
	}
	logrus.Debugf("Giving new service the following IP: %v", freeIp.String())

	// The user won't know the IP address, so we'll need to replace all the IP address placeholders with the actual
	//  IP
	replacedStartCmd, replacedEnvVars := replaceIpPlaceholderForDockerParams(
		ipPlaceholder,
		freeIp,
		startCmd,
		dockerEnvVars)

	portBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		portBindings[port] = nil
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		context,
		imageName,
		launcher.dockerNetworkId,
		freeIp,
		map[commons.ContainerCapability]bool{},
		commons.DefaultNetworkMode,
		portBindings,
		replacedStartCmd,
		replacedEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{
			launcher.testVolumeName: testVolumeMountFilepath,
		},
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", imageName)
	}
	return containerId, freeIp, nil
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
