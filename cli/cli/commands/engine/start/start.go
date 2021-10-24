package start

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	engine_labels_schema2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_labels_schema"
	output_printers2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	CommandStr = "start"

	engineImageArg = "image"

	containerNamePrefix = "kurtosis-engine"

	networkToStartEngineContainerIn = "bridge"

	dockerSocketFilepath = "/var/run/docker.sock"

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	engineWaitForReadyTimeout = 10 * time.Second

	engineImageInfoLabel = "Image"
	engineApiVersionInfoLabel = "API Version"
)

var engineImage string

var StartCmd = &cobra.Command{
	Use:   CommandStr,
	Short: "Starts the Kurtosis engine",
	Long: "Starts the Kurtosis engine, doing nothing if an engine is already running",
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

	// Don't start an engine if one is already running
	matchingEngineContainers, err := dockerManager.GetContainersByLabels(
		ctx,
		engine_labels_schema2.EngineContainerLabels,
		shouldGetStoppedContainersWhenCheckingForExistingEngines,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting existing engine containers")
	}
	numMatchingEngineContainers := len(matchingEngineContainers)
	if numMatchingEngineContainers > 1 {
		return stacktrace.NewError("Found %v running engine containers; there should never be more than 1 engine container!", numMatchingEngineContainers)
	}
	if numMatchingEngineContainers > 0 {
		logrus.Info("A Kurtosis engine is already running; nothing to do")
		return nil
	}

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
		engine_labels_schema2.EngineContainerLabels,
	).Build()

	_, hostMachinePortBindings, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	hostMachineEnginePortBinding, found := hostMachinePortBindings[enginePortObj]
	if !found {
		return stacktrace.NewError("The Kurtosis engine server started successfully, but no host machine port binding was found")
	}

	engineInfo, err := waitUntilAvailableAndGetEngineInfo(ctx, hostMachineEnginePortBinding)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying that the engine is up by getting the engine info")
	}

	engineInfoPrinter := output_printers2.NewKeyValuePrinter()
	engineInfoPrinter.AddPair(engineImageInfoLabel, engineImage)
	engineInfoPrinter.AddPair(engineApiVersionInfoLabel, engineInfo.EngineApiVersion)

	logrus.Info("Kurtosis engine started successfully with the following info:")
	engineInfoPrinter.Print()

	return nil
}

// NOTE: We can't replace this with the higher-level API because we need to set the WaitForReady flag
func waitUntilAvailableAndGetEngineInfo(ctx context.Context, hostMachineEnginePortBinding *nat.PortBinding) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, engineWaitForReadyTimeout)
	defer cancelFunc()
	engineUrl := fmt.Sprintf("%v:%v", hostMachineEnginePortBinding.HostIP, hostMachineEnginePortBinding.HostPort)
	conn, err := grpc.Dial(engineUrl, grpc.WithInsecure())
	engineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	engineInfo, err := engineClient.GetEngineInfo(ctxWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for %v for the engine to become available and return engine info",
			engineWaitForReadyTimeout,
		)
	}
	return engineInfo, nil

}