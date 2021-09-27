package api_container_starter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core-launcher-lib/lib/api_container_docker_consts"
	"github.com/palantir/stacktrace"
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
		listenPort uint,
		listenProtocol string,
		networkId string,
		ipAddr net.IP,
		enclaveDataVolName string,
		args APIContainerArgs,
) (string, error) {
	if err := args.Validate(); err != nil {
		return "", stacktrace.Propagate(err, "Can't start API container because args didn't pass validation")
	}

	kurtosisApiPort := nat.Port(fmt.Sprintf(
		"%v/%v",
		listenPort,
		listenProtocol,
	))

	serializedArgsBytes, err := json.Marshal(args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing API container args to JSON")
	}
	serializedArgsStr := string(serializedArgsBytes)

	envVars := map[string]string{
		api_container_docker_consts.SerializedArgsEnvVar: serializedArgsStr,
	}

	// For now, we don't publish the API container's port to the host machine (though maybe this will change in the future)
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(map[nat.Port]bool{
		kurtosisApiPort: true,
	}).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(map[string]string{
		dockerSocket: dockerSocket,
	}).WithVolumeMounts(map[string]string{
		enclaveDataVolName: api_container_docker_consts.EnclaveDataVolumeMountpoint,
	}).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred starting the API container")
	}

	return containerId, nil
}
