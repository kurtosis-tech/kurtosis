package engine_manager

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/start"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/stop"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
	"path"
	"time"
)

type EngineStatus string
const (
	EngineStatus_Stopped                                EngineStatus = "STOPPED"
	EngineStatus_ContainerRunningButServerNotResponding EngineStatus = "CONTAINER_RUNNING_BUT_SERVER_NOT_RESPONDING"
	EngineStatus_Running                                EngineStatus = "RUNNING"

	waitForEngineResponseTimeout = 5 * time.Second
	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	localHostIPAddressStr = "0.0.0.0"
)

type EngineManager struct {
	// Make engine IP & port configurable in the future
}

func NewEngineManager() *EngineManager {
	return &EngineManager{}
}

// NOTE: The first second value, the engine API version, will only be filled in if the engine status is "running"
func GetEngineStatus(ctx context.Context, dockerManager *docker_manager.DockerManager) (EngineStatus, string, error) {
	runningEngineContainers, err := dockerManager.GetContainersByLabels(ctx, engine_labels_schema.EngineContainerLabels, shouldGetStoppedContainersWhenCheckingForExistingEngines)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return "", "", stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	}
	if numRunningEngineContainers == 0 {
		return EngineStatus_Stopped, "", nil
	}
	engineContainer := runningEngineContainers[0]

	enginePortObj, err := nat.NewPort(
		kurtosis_engine_rpc_api_consts.ListenProtocol,
		fmt.Sprintf("%v", kurtosis_engine_rpc_api_consts.ListenPort),
	)
	if err != nil {
		return "", "", stacktrace.Propagate(
			err,
			"An error occurred creating an engine port object from port num '%v' and protocol '%v'",
			kurtosis_engine_rpc_api_consts.ListenPort,
			kurtosis_engine_rpc_api_consts.ListenProtocol,
		)
	}

	hostMachineEnginePortBinding, found := engineContainer.GetHostPortBindings()[enginePortObj]
	if !found {
		return "", "", stacktrace.NewError("Found a Kurtosis engine server container, but it didn't have a host machine port binding - this is likely a Kurtosis bug")
	}

	engineInfo, err := getEngineInfoWithTimeout(ctx, hostMachineEnginePortBinding)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, "", nil
	}

	return EngineStatus_Running, engineInfo.EngineApiVersion, nil
}

// Gets an engine client connected to the local engine
// If no engine is running, attempts to start one first
func GetEngineClient(ctx context.Context, dockerManager *docker_manager.DockerManager) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	// Check the engine status first so we can print a helpful message in case the engine isn't running
	status, _, err := GetEngineStatus(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
	binaryFilename := path.Base(os.Args[0])
	switch status {
	case EngineStatus_Stopped:
		return nil, nil, stacktrace.NewError(
			"No Kurtosis engine is running; you'll need to start one by running '%v %v %v'",
			binaryFilename,
			engine.CommandStr,
			start.CommandStr,
		)
	case EngineStatus_ContainerRunningButServerNotResponding:
		return nil, nil, stacktrace.NewError(
			"A Kurtosis engine container is running, but it's not responding; this shouldn't happen and you'll likely " +
				"want to restart the engine by running '%v %v %v && %v %v %v'",
			binaryFilename,
			engine.CommandStr,
			stop.CommandStr,
			binaryFilename,
			engine.CommandStr,
			start.CommandStr,
		)
	case EngineStatus_Running:
		// This is the happy case; nothing to do
	default:
		return nil, nil, stacktrace.NewError("Unrecognized engine status '%v'; this is a bug in Kurtosis", status)
	}

	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, kurtosis_engine_rpc_api_consts.ListenPort)

	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	conn, err := grpc.Dial(kurtosisEngineSocketStr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Engine Server at '%v'",
			kurtosisEngineSocketStr,
		)
	}

	engineServiceClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)

	return engineServiceClient, conn.Close, nil
}

// TODO StopEngine

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
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

