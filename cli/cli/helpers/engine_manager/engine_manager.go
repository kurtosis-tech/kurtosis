package engine_manager

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"strings"
	"time"
)

const (

	waitForEngineResponseTimeout = 5 * time.Second
	shouldGetStoppedContainersWhenCheckingForExistingEngines = false

	engineStopTimeout = 30 * time.Second
)

var engineContainerLabels = map[string]string{
	forever_constants.AppIDLabel: forever_constants.AppIDValue,
	schema.ContainerTypeLabel: schema.ContainerTypeEngineServer,
}

type EngineManager struct {
	dockerManager *docker_manager.DockerManager
	// Make engine IP, port, and protocol configurable in the future

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewEngineManager(dockerManager *docker_manager.DockerManager, objAttrsProvider schema.ObjectAttributesProvider) *EngineManager {
	return &EngineManager{dockerManager: dockerManager, objAttrsProvider: objAttrsProvider}
}

/*
Returns:
	- The engine status
	- The host machine port bindings (not present if the engine is stopped)
	- The engine version (only present if the engine is running)
 */
func (manager *EngineManager) GetEngineStatus(
	ctx context.Context,
) (EngineStatus, *nat.PortBinding, string, error) {
	runningEngineContainers, err := manager.dockerManager.GetContainersByLabels(ctx, engineContainerLabels, shouldGetStoppedContainersWhenCheckingForExistingEngines)
	if err != nil {
		return "", nil, "", stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers")
	}

	numRunningEngineContainers := len(runningEngineContainers)
	if numRunningEngineContainers > 1 {
		return "", nil, "", stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engine containers; this is very strange as there should never be more than one", numRunningEngineContainers)
	}
	if numRunningEngineContainers == 0 {
		return EngineStatus_Stopped, nil, "", nil
	}
	engineContainer := runningEngineContainers[0]

	engineContainerLabels := engineContainer.GetLabels()
	enginePortNumStr, found := engineContainerLabels[schema.PortNumLabel]
	if !found {
		return "", nil, "", stacktrace.NewError("Found running engine container, but it didn't have label '%v'", schema.PortNumLabel)
	}
	enginePortProtocol, found := engineContainerLabels[schema.PortProtocolLabel]
	if !found {
		return "", nil, "", stacktrace.NewError("Found running engine container, but it didn't have label '%v'", schema.PortProtocolLabel)
	}

	enginePortObj, err := nat.NewPort(
		enginePortProtocol,
		enginePortNumStr,
	)
	if err != nil {
		return "", nil, "", stacktrace.Propagate(
			err,
			"An error occurred creating an engine port object from port num '%v' and protocol '%v'",
			enginePortNumStr,
			enginePortProtocol,
		)
	}

	hostMachineEnginePortBinding, found := engineContainer.GetHostPortBindings()[enginePortObj]
	if !found {
		return "", nil, "", stacktrace.NewError("Found a Kurtosis engine server container, but it didn't have a host machine port binding - this is likely a Kurtosis bug")
	}

	engineClient, clientCloseFunc, err := getEngineClientFromHostPortBinding(hostMachineEnginePortBinding)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, hostMachineEnginePortBinding, "", nil
	}
	defer clientCloseFunc()

	engineInfo, err := getEngineInfoWithTimeout(ctx, engineClient)
	if err != nil {
		return EngineStatus_ContainerRunningButServerNotResponding, hostMachineEnginePortBinding, "", nil
	}

	return EngineStatus_Running, hostMachineEnginePortBinding, engineInfo.GetEngineVersion(), nil
}

// Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithDefaultVersion(ctx context.Context, logLevel logrus.Level) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	status, maybeHostMachinePortBinding, engineVersion, err := manager.GetEngineStatus(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
	engineGuarantor := newEngineExistenceGuarantorWithDefaultVersion(
		ctx,
		maybeHostMachinePortBinding,
		manager.dockerManager,
		manager.objAttrsProvider,
		logLevel,
		engineVersion,
	)
	engineClient, engineClientCloseFunc, err := startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}

// Starts an engine if one doesn't exist already, and returns a client to it
func (manager *EngineManager) StartEngineIdempotentlyWithCustomVersion(ctx context.Context, engineImageVersionTag string, logLevel logrus.Level) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	status, maybeHostMachinePortBinding, engineVersion, err := manager.GetEngineStatus(ctx)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
	engineGuarantor := newEngineExistenceGuarantorWithCustomVersion(
		ctx,
		maybeHostMachinePortBinding,
		manager.dockerManager,
		manager.objAttrsProvider,
		engineImageVersionTag,
		logLevel,
		engineVersion,
	)
	engineClient, engineClientCloseFunc, err := startEngineWithGuarantor(ctx, status, engineGuarantor)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting the engine with the engine existence guarantor")
	}
	return engineClient, engineClientCloseFunc, nil
}


// Stops the engine if it's running, doing nothing if not
func (manager *EngineManager) StopEngineIdempotently(ctx context.Context) error {
	matchingEngineContainers, err := manager.dockerManager.GetContainersByLabels(
		ctx,
		engineContainerLabels,
		shouldGetStoppedContainersWhenCheckingForExistingEngines,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting Kurtosis engine containers, which we need to check for existing engines")
	}

	numMatchingEngineContainers := len(matchingEngineContainers)
	if numMatchingEngineContainers == 0 {
		return nil
	}
	if numMatchingEngineContainers > 1 {
		logrus.Warnf(
			"Found %v Kurtosis engine containers, which is strange because there should never be more than 1 engine container; all will be stopped",
			numMatchingEngineContainers,
		)
	}

	engineStopErrorStrs := []string{}
	for _, engineContainer := range matchingEngineContainers {
		containerName := engineContainer.GetName()
		containerId := engineContainer.GetId()
		if err := manager.dockerManager.StopContainer(ctx, containerId, engineStopTimeout); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping engine container '%v' with ID '%v'",
				containerName,
				containerId,
			)
			engineStopErrorStrs = append(engineStopErrorStrs, wrappedErr.Error())
		}
	}

	if len(engineStopErrorStrs) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred stopping the engine(s):\n%v",
			strings.Join(
				engineStopErrorStrs,
				"\n\n",
			),
		)
	}
	return nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func startEngineWithGuarantor(ctx context.Context, currentStatus EngineStatus, engineGuarantor *engineExistenceGuarantor) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	if err := currentStatus.Accept(engineGuarantor); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred guaranteeing that a Kurtosis engine is running")
	}
	hostMachinePortBinding := engineGuarantor.getPostVisitingHostMachinePortBinding()

	engineClient, clientCloseFunc, err := getEngineClientFromHostPortBinding(hostMachinePortBinding)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to the running engine; this is very strange and likely indicates a bug in the engine itself")
	}

	// Final verification to ensure that the engine server is responding
	if _, err := getEngineInfoWithTimeout(ctx, engineClient); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to the engine server; this is very strange and likely indicates a bug in the engine itself")
	}
	return engineClient, clientCloseFunc, nil
}

func getEngineClientFromHostPortBinding(hostMachinePortBinding *nat.PortBinding) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	url := fmt.Sprintf("%v:%v", hostMachinePortBinding.HostIP, hostMachinePortBinding.HostPort)
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling Kurtosis engine at URL '%v'", url)
	}
	engineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	return engineClient, conn.Close, nil
}

func getEngineInfoWithTimeout(ctx context.Context, client kurtosis_engine_rpc_api_bindings.EngineServiceClient) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	ctxWithTimeout, cancelFunc := context.WithTimeout(ctx, waitForEngineResponseTimeout)
	defer cancelFunc()
	engineInfo, err := client.GetEngineInfo(ctxWithTimeout, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Kurtosis engine server didn't return a response even with %v timeout",
			waitForEngineResponseTimeout,
		)
	}
	return engineInfo, nil
}
