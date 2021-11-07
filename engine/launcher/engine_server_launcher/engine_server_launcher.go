/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	dockerSocketFilepath = "/var/run/docker.sock"

	// TODO Move this into the ObjAttrProvider schema
	containerNamePrefix = "kurtosis-engine"

	networkToStartEngineContainerIn = "bridge"

	listenProtocol = "tcp"

	// The location where the engine data directory (on the Docker host machine) will be bind-mounted
	//  on the engine server
	EngineDataDirpathOnEngineServerContainer = "/engine-data"

	maxWaitForAvailabilityRetries         = 10
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	availabilityWaitingExecCmdSuccessExitCode = 0

	// TODO This needs to be merged into the obj attrs schema lib!!!!!!
	ContainerTypeKurtosisEngine = "kurtosis-engine"
)

// TODO This should be pushed into the obj attributes schema lib!!!!!
var EngineContainerLabels = map[string]string{
	// TODO These need refactoring!!! "ContainerTypeLabel" and "AppIDLabel" aren't just for enclave objects!!!
	//  See https://github.com/kurtosis-tech/kurtosis-cli/issues/24
	forever_constants.AppIDLabel:         forever_constants.AppIDValue,
	schema.ContainerTypeLabel: ContainerTypeKurtosisEngine,
}

type EngineServerLauncher struct {
	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewEngineServerLauncher(dockerManager *docker_manager.DockerManager, objAttrsProvider schema.ObjectAttributesProvider) *EngineServerLauncher {
	return &EngineServerLauncher{dockerManager: dockerManager, objAttrsProvider: objAttrsProvider}
}

func (launcher *EngineServerLauncher) Launch(
	ctx context.Context,
	containerImage string,
	logLevel logrus.Level,
	listenPortNum uint16, // The port that the engine server will listen on AND the port that it should be bound to on the host machine
	engineDataDirpathOnHostMachine string,
) (*nat.PortBinding, error) {
	matchingNetworks, err := launcher.dockerManager.GetNetworksByName(ctx, networkToStartEngineContainerIn)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			networkToStartEngineContainerIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return nil, stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			networkToStartEngineContainerIn,
			numMatchingNetworks,
		)
	}
	targetNetwork := matchingNetworks[0]
	targetNetworkId := targetNetwork.GetId()

	containerStartTimeUnixSecs := time.Now().Unix()
	containerName := fmt.Sprintf(
		"%v_%v",
		containerNamePrefix,
		containerStartTimeUnixSecs,
	)
	enginePortObj, err := nat.NewPort(
		listenProtocol,
		fmt.Sprintf("%v", listenPortNum),
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a port object with port num '%v' and protocol '%v' to represent the engine's port",
			listenPortNum,
			listenProtocol,
		)
	}

	argsObj, err := args.NewEngineServerArgs(
		listenPortNum,
		listenProtocol,
		logLevel.String(),
		engineDataDirpathOnHostMachine,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine server args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		enginePortObj: docker_manager.NewManualPublishingSpec(listenPortNum),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath:           dockerSocketFilepath,
		engineDataDirpathOnHostMachine: EngineDataDirpathOnEngineServerContainer,
	}
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		targetNetworkId,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		EngineContainerLabels,
	).Build()

	containerId, hostMachinePortBindings, err := launcher.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	shouldKillEngineContainer := true
	defer func() {
		if shouldKillEngineContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf("Launching the engine server didn't complete successfully so we tried to kill the container we started, but doing so exited with an error:\n%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill engine server with container ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := waitForAvailability(ctx, launcher.dockerManager, containerId, listenPortNum); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the engine server to become available")
	}

	hostMachineEnginePortBinding, found := hostMachinePortBindings[enginePortObj]
	if !found {
		return nil, stacktrace.NewError("The Kurtosis engine server started successfully, but no host machine port binding was found")
	}

	shouldKillEngineContainer = false
	return hostMachineEnginePortBinding, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func waitForAvailability(ctx context.Context, dockerManager *docker_manager.DockerManager, containerId string, listenPortNum uint16) error {
	commandStr := fmt.Sprintf(
		"[ -n \"$(netstat -anp %v | grep LISTEN | grep %v)\" ]",
		listenProtocol,
		listenPortNum,
	)
	execCmd := []string{
		"sh",
		"-c",
		commandStr,
	}
	for i := 0; i < maxWaitForAvailabilityRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if (exitCode == availabilityWaitingExecCmdSuccessExitCode) {
				return nil
			}
			logrus.Debugf(
				"Engine server availability-waiting command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				availabilityWaitingExecCmdSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Engine server availability-waiting command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxWaitForAvailabilityRetries {
			time.Sleep(timeBetweenWaitForAvailabilityRetries)
		}
	}

	return stacktrace.NewError(
		"The engine server didn't become available (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxWaitForAvailabilityRetries,
		timeBetweenWaitForAvailabilityRetries,
	)
}
