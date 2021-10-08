package api_container_starter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_manager/api_container_launcher_lib/api_container_docker_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	dockerSocket = "/var/run/docker.sock"
)

// This is a helper function that should be usable by ALL launcher versions, because the parameters are very unlikely
//  to change now or ever
func StartAPIContainer(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		containerImage string,
		containerName string,
		containerLabels map[string]string,
		listenPort uint,
		listenProtocol string,
		networkId string,
		ipAddr net.IP,
		enclaveDataVolName string,
		args APIContainerArgs,
) (string, *nat.PortBinding, error) {
	if err := args.Validate(); err != nil {
		return "", nil, stacktrace.Propagate(err, "Can't start API container because args didn't pass validation")
	}

	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		listenPort,
		listenProtocol,
	))

	serializedArgsBytes, err := json.Marshal(args)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred serializing API container args to JSON")
	}
	serializedArgsStr := string(serializedArgsBytes)

	envVars := map[string]string{
		api_container_docker_consts.SerializedArgsEnvVar: serializedArgsStr,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(map[nat.Port]bool{
		kurtosisApiPort: true,
	}).ShouldPublishAllPorts(
		true,	// We always publish the API container's ports so that we can call its external container registration functions from the CLI
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(map[string]string{
		dockerSocket: dockerSocket,
	}).WithVolumeMounts(map[string]string{
		enclaveDataVolName: api_container_docker_consts.EnclaveDataVolumeMountpoint,
	}).WithLabels(
		containerLabels,
	).Build()

	containerId, hostPortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the API container")
	}
	shouldDeleteContainer := true
	defer func() {
		if shouldDeleteContainer {
			if killErr := dockerManager.KillContainer(context.Background(), containerId); killErr != nil {
				logrus.Errorf("The function to create the API container didn't finish successful so we tried to kill the container we created, but the killing threw an error:")
				logrus.Error(killErr)
			}
		}
	}()

	hostPortBinding, found := hostPortBindings[kurtosisApiPort]
	if !found {
		return "", nil, stacktrace.NewError("No host port binding was found for API container port '%v' - this is very strange!", kurtosisApiPort)
	}

	shouldDeleteContainer = false
	return containerId, hostPortBinding, nil
}
