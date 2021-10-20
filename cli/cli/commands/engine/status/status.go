package status

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-cli/cli/output_printers"
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
	engineStoppedStatusStr = "Kurtosis engine is not running"
	engineContainerRunningButServerNotRespondingStatusStr = "A Kurtosis engine container is running, but the server inside couldn't be reached"
	engineRunningStatusStrHeaderLine = "Kurtosis engine is running with the following info:"

	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	waitForEngineResponseTimeout = 5 * time.Second

	engineApiVersionInfoLabel = "API Version"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Reports the status of the Kurtosis engine",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	runningEngineContainers, err := dockerManager.GetContainersByLabels(ctx, engine_labels_schema.EngineContainerLabels, shouldGetStoppedContainersWhenCheckingForExistingEngines)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	if err := printEngineStatusUsingRunningContainers(ctx, runningEngineContainers); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the Kurtosis engine status")
	}
	return nil
}

func printEngineStatusUsingRunningContainers(ctx context.Context, runningEngineContainers []*types.Container) error {
	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	}
	if numRunningEngineContainers == 0 {
		fmt.Fprintln(logrus.StandardLogger().Out, engineStoppedStatusStr)
		return nil
	}
	engineContainer := runningEngineContainers[0]

	enginePortObj, err := nat.NewPort(
		kurtosis_engine_rpc_api_consts.ListenProtocol,
		fmt.Sprintf("%v", kurtosis_engine_rpc_api_consts.ListenPort),
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating an engine port object from port num '%v' and protocol '%v'",
			kurtosis_engine_rpc_api_consts.ListenPort,
			kurtosis_engine_rpc_api_consts.ListenProtocol,
		)
	}

	hostMachineEnginePortBinding, found := engineContainer.GetHostPortBindings()[enginePortObj]
	if !found {
		return stacktrace.NewError("Found a Kurtosis engine server container, but it didn't have a host machine port binding - this is likely a Kurtosis bug")
	}

	engineInfo, err := getEngineInfoWithTimeout(ctx, hostMachineEnginePortBinding)
	if err != nil {
		fmt.Fprintln(logrus.StandardLogger().Out, engineContainerRunningButServerNotRespondingStatusStr)
		return nil
	}

	engineInfoPrinter := output_printers.NewKeyValuePrinter()
	engineInfoPrinter.AddPair(engineApiVersionInfoLabel, engineInfo.EngineApiVersion)

	fmt.Fprintln(logrus.StandardLogger().Out, engineRunningStatusStrHeaderLine)
	engineInfoPrinter.Print()

	return nil
}

// NOTE: We can't replace this with the higher-level API because we need to set the WaitForReady flag
func getEngineInfoWithTimeout(ctx context.Context, hostMachineEnginePortBinding *nat.PortBinding) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForEngineResponseTimeout)
	defer cancelFunc()
	engineUrl := fmt.Sprintf("%v:%v", hostMachineEnginePortBinding.HostIP, hostMachineEnginePortBinding.HostPort)
	conn, err := grpc.Dial(engineUrl, grpc.WithInsecure())
	engineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	engineInfo, err := engineClient.GetEngineInfo(ctxWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Kurtosis engine server didn't return a response even with %v timeout",
			waitForEngineResponseTimeout,
		)
	}
	return engineInfo, nil
}
