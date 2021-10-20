package start

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

const (
	engineImageArg = "image"

	containerNamePrefix = "kurtosis-engine"

	networkToStartEngineContainerIn = "bridge"

	dockerSocketFilepath = "/var/run.docker.sock"
)

var engineImage string

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the Kurtosis engine",
	RunE:  run,
}

func init() {
	StartCmd.Flags().StringVar(
		&engineImage,
		engineImageArg,
		defaults.DefaultEngineImage,
		"The image of the Kurtosis engine that should be started",
	)
}

func run(cmd *cobra.Command, args []string) error {
	logrus.Infof("Starting Kurtosis engine from image '%v'...", engineImage)

	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	matchingNetworks, err := dockerManager.GetNetworksByName(ctx, networkToStartEngineContainerIn)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			networkToStartEngineContainerIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return stacktrace.NewError(
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
		kurtosis_engine_rpc_api_consts.ListenProtocol,
		fmt.Sprintf("%v", kurtosis_engine_rpc_api_consts.ListenPort),
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating a port object with port num '%v' and protocol '%v' to represent the engine's port",
			kurtosis_engine_rpc_api_consts.ListenPort,
			kurtosis_engine_rpc_api_consts.ListenProtocol,
		)
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		enginePortObj: docker_manager.NewManualPublishingSpec(kurtosis_engine_rpc_api_consts.ListenPort),
	}
	labels := map[string]string{
		// TODO These need refactoring!!! "ContainerTypeLabel" and "AppIDLabel" aren't just for enclave objects!!!
		//  See https://github.com/kurtosis-tech/kurtosis-cli/issues/24
		enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
		enclave_object_labels.ContainerTypeLabel: engine_labels_schema.ContainerTypeKurtosisEngine,
	}
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		engineImage,
		containerName,
		targetNetworkId,
	).WithBindMounts(map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath: dockerSocketFilepath,
	}).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labels,
	).Build()

	if _, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}

	logrus.Info("Kurtosis engine started successfully")
	return nil
}