package engine_status_retriever

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type EngineStatus string
const (
	EngineStatus_Stopped EngineStatus = "STOPPED"
	EngineStatus_ContainerRunningButServerNotResponding EngineStatus = "CONTAINER_RUNNING_BUT_SERVER_NOT_RESPONDING"
	EngineStatus_Running EngineStatus = "RUNNING"

	waitForEngineResponseTimeout = 5 * time.Second
	shouldGetStoppedContainersWhenCheckingForExistingEngines = false
)

// NOTE: The first second value, the engine API version, will only be filled in if the engine status is "running"
func RetrieveEngineStatus(ctx context.Context, dockerManager *docker_manager.DockerManager) (EngineStatus, string, error) {
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
		return 0, "", stacktrace.Propagate(
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
