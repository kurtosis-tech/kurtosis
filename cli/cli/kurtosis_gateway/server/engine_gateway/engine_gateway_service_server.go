package engine_gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_gateway/connection"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_gateway/live_engine_client_supplier"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_gateway/port_utils"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_gateway/run/api_container_gateway"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_gateway/server/common"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	restclient "k8s.io/client-go/rest"
)

const (
	// API Container gateways spun up by this engine gateway run on localhost
	localHostIpStr = "127.0.0.1"

	waitForGatewayGrpcReady = true

	apiContainerGatewayHealthcheckTimeout = 5 * time.Second
)

type EngineGatewayServiceServer struct {
	// Client for the engine we'll be connecting too
	engineClientSupplier *live_engine_client_supplier.LiveEngineClientSupplier

	// Configuration for the kubernetes cluster kurtosis is running on
	kubernetesConfig *restclient.Config

	connectionProvider *connection.GatewayConnectionProvider

	mutex                        *sync.Mutex
	enclaveIdToRunningGatewayMap map[string]*runningApiContainerGateway
}

type runningApiContainerGateway struct {
	// info about the api container on the host machine
	hostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	// closeGatewayFunc
	closeFunc func()
}

// NewEngineGatewayServiceServer returns a EngineGatewayServiceServer
// runningEngine is a kurtosis engine running a cluster that can be reached through clients configured with kubernetesConfig
func NewEngineGatewayServiceServer(connectionProvider *connection.GatewayConnectionProvider, engineClientSupplier *live_engine_client_supplier.LiveEngineClientSupplier) (resultGatewayService *EngineGatewayServiceServer, gatewayCloseFunc func()) {
	// We start out with no enclave api-container gateways runnings
	runningApiContainers := map[string]*runningApiContainerGateway{}

	service := &EngineGatewayServiceServer{
		engineClientSupplier:         engineClientSupplier,
		kubernetesConfig:             nil,
		connectionProvider:           connectionProvider,
		mutex:                        &sync.Mutex{},
		enclaveIdToRunningGatewayMap: runningApiContainers,
	}
	closeFunc := func() {
		// Kill the running enclave gateways
		for enclaveId := range service.enclaveIdToRunningGatewayMap {
			service.idempotentKillRunningGatewayForEnclaveId(enclaveId)
		}
	}

	return service, closeFunc
}

func (service *EngineGatewayServiceServer) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	remoteEngineResponse, err := remoteEngineClient.CreateEnclave(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave through the remote engine")
	}
	cleanUpEnclave := true
	defer func() {
		if cleanUpEnclave {
			destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveIdentifier: args.GetEnclaveName()}
			if _, err := remoteEngineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
				logrus.Error("Launching the Enclave gateway failed, expected to be able to cleanup the created enclave, but an error occurred calling the backend to destroy the enclave we created:")
				//out.PrintErrLn(err.Error())
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill the enclave with name '%v'", args.GetEnclaveName())
			}
		}
	}()
	// Update enclave API Container Host Machine Info
	// We want to update the host machine info for the api container
	createdEnclaveInfo := remoteEngineResponse.GetEnclaveInfo()
	createdEnclaveId := createdEnclaveInfo.GetEnclaveUuid()

	runningApiContainerGateway, err := service.startRunningGatewayForEnclave(createdEnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to start a local gateway for enclave '%v', instead a non-nil err was returned", createdEnclaveId)
	}
	cleanupRunningApiContainerGateway := true
	defer func() {
		if cleanupRunningApiContainerGateway {
			service.idempotentKillRunningGatewayForEnclaveId(createdEnclaveId)
		}
	}()

	// Overwrite the hostmachineinfo for the apicontainer returned by the remote engine
	if remoteEngineResponse.EnclaveInfo == nil {
		return nil, stacktrace.NewError("Expected the response from the remote engine to have info on the enclave '%v', instead no enclave information was found.", createdEnclaveId)
	}
	remoteEngineResponse.EnclaveInfo.ApiContainerHostMachineInfo = runningApiContainerGateway.hostMachineInfo

	cleanUpEnclave = false
	cleanupRunningApiContainerGateway = false
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetEnclaves(ctx context.Context, in *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	remoteEngineResponse, err := remoteEngineClient.GetEnclaves(ctx, in)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves from the remote engine")
	}
	responseEnclaves := remoteEngineResponse.GetEnclaveInfo()
	cleanUpRunningGateways := true
	for enclaveId, enclaveInfo := range responseEnclaves {
		var runningApiContainerGateway *runningApiContainerGateway
		runningApiContainerGateway, isRunning := service.enclaveIdToRunningGatewayMap[enclaveId]
		// If the gateway isn't running, start it
		if !isRunning {
			runningApiContainerGateway, err = service.startRunningGatewayForEnclave(enclaveInfo)
			defer func() {
				if cleanUpRunningGateways {
					service.idempotentKillRunningGatewayForEnclaveId(enclaveId)
				}
			}()
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to start a local gateway for enclave '%v', instead a non-nil error was returned", enclaveId)
			}
		}
		remoteEngineResponse.EnclaveInfo[enclaveId].ApiContainerHostMachineInfo = runningApiContainerGateway.hostMachineInfo
	}

	cleanUpRunningGateways = false
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetEnclavesByUuids(ctx context.Context, in *kurtosis_engine_rpc_api_bindings.GetEnclavesByUuidsArgs) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	remoteEngineResponse, err := remoteEngineClient.GetEnclavesByUuids(ctx, in)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves from the remote engine")
	}
	responseEnclaves := remoteEngineResponse.GetEnclaveInfo()
	cleanUpRunningGateways := true
	for enclaveId, enclaveInfo := range responseEnclaves {
		var runningApiContainerGateway *runningApiContainerGateway
		runningApiContainerGateway, isRunning := service.enclaveIdToRunningGatewayMap[enclaveId]
		// If the gateway isn't running, start it
		if !isRunning {
			runningApiContainerGateway, err = service.startRunningGatewayForEnclave(enclaveInfo)
			defer func() {
				if cleanUpRunningGateways {
					service.idempotentKillRunningGatewayForEnclaveId(enclaveId)
				}
			}()
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to start a local gateway for enclave '%v', instead a non-nil error was returned", enclaveId)
			}
		}
		remoteEngineResponse.EnclaveInfo[enclaveId].ApiContainerHostMachineInfo = runningApiContainerGateway.hostMachineInfo
	}

	cleanUpRunningGateways = false
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetExistingAndHistoricalEnclaveIdentifiers(ctx context.Context, args *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetExistingAndHistoricalEnclaveIdentifiersResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}

	response, err := remoteEngineClient.GetExistingAndHistoricalEnclaveIdentifiers(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred calling remote engine to get historical identifiers enclave")
	}

	return response, nil
}

func (service *EngineGatewayServiceServer) StopEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.StopEnclaveArgs) (*emptypb.Empty, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	if _, err := remoteEngineClient.StopEnclave(ctx, args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred calling remote engine to stop enclave '%v'", args.EnclaveIdentifier)
	}

	return &emptypb.Empty{}, nil
}

func (service *EngineGatewayServiceServer) DestroyEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs) (*emptypb.Empty, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	if _, err := remoteEngineClient.DestroyEnclave(ctx, args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred calling remote engine to destroy enclave with ID '%v':", args.EnclaveIdentifier)
	}
	// Kill the running api container gateway for this enclave
	enclaveIdOfGatewayToKill := args.EnclaveIdentifier
	service.idempotentKillRunningGatewayForEnclaveId(enclaveIdOfGatewayToKill)

	return &emptypb.Empty{}, nil
}

func (service *EngineGatewayServiceServer) Clean(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CleanArgs) (*kurtosis_engine_rpc_api_bindings.CleanResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	remoteEngineResponse, err := remoteEngineClient.Clean(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}
	gatewaysToKill := remoteEngineResponse.RemovedEnclaveNameAndUuids
	for _, gatewayEnclaveNameAndUuid := range gatewaysToKill {
		service.idempotentKillRunningGatewayForEnclaveId(gatewayEnclaveNameAndUuid.Uuid)
	}
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetEngineInfo(ctx context.Context, emptyArgs *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	remoteEngineResponse, err := remoteEngineClient.GetEngineInfo(ctx, emptyArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine info through the remote engine")
	}
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetServiceLogs(
	args *kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs,
	streamToWriteTo kurtosis_engine_rpc_api_bindings.EngineService_GetServiceLogsServer,
) error {
	remoteEngineClient, err := service.engineClientSupplier.GetEngineClient()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get a client for a live Kurtosis engine, instead a non nil error was returned")
	}
	streamToReadFrom, err := remoteEngineClient.GetServiceLogs(streamToWriteTo.Context(), args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service logs")
	}
	if err := common.ForwardKurtosisExecutionStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse](streamToReadFrom, streamToWriteTo); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from Kurtosis engine back to the user")
	}
	return nil
}

// Private functions for managing our running enclave api container gateways
func (service *EngineGatewayServiceServer) startRunningGatewayForEnclave(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (*runningApiContainerGateway, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	enclaveId := enclaveInfo.GetEnclaveUuid()
	// Ask the kernel for a free open TCP port
	gatewayPortSpec, err := port_utils.GetFreeTcpPort(localHostIpStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a free, open TCP port on host, instead a non-nil error was returned")
	}
	// Channel for messages to stop the running server
	gatewayStopChannel := make(chan struct{}, 1)

	// Info for how to connect to the api container through the gateway running on host machine
	apiContainerHostMachineInfo := &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
		IpOnHostMachine:       localHostIpStr,
		GrpcPortOnHostMachine: uint32(gatewayPortSpec.GetNumber()),
		// TODO proxy endpoint for gateway
	}

	// Start the server in a goroutine
	// Stop the running gateway
	gatewayStopFunc := func() {
		// Send message to stop the gateway server
		gatewayStopChannel <- struct{}{}
	}
	// TODO: Modify MinimalGrpcServer.RunUntilStopped to take in a `ReadyChannel` to communicate when a GRPC server is ready to serve
	// Currently, we have to make a health check request to verify that the API container gateway is ready
	go func() {
		if err := api_container_gateway.RunApiContainerGatewayUntilStopped(service.connectionProvider, enclaveInfo, gatewayPortSpec.GetNumber(), gatewayStopChannel); err != nil {
			logrus.Warnf("Expected to run api container gateway until stopped, but the server exited prematurely with a non-nil error: '%v'", err)
		}
	}()
	cleanUpGateway := true
	defer func() {
		if cleanUpGateway {
			gatewayStopFunc()
		}
	}()
	// Need to wait for the GRPC server spun up in the goFunc to be ready
	if err := waitForGatewayReady(apiContainerHostMachineInfo); err != nil {
		logrus.Errorf("Expected Gateway to be reachable, instead an error was returned:\n%v", err)
	}

	runningGatewayInfo := &runningApiContainerGateway{
		closeFunc:       gatewayStopFunc,
		hostMachineInfo: apiContainerHostMachineInfo,
	}
	// Store information about our running gateway
	service.enclaveIdToRunningGatewayMap[enclaveId] = runningGatewayInfo
	cleanUpMapEntry := true
	defer func() {
		if cleanUpMapEntry {
			delete(service.enclaveIdToRunningGatewayMap, enclaveId)
		}
	}()

	cleanUpMapEntry = false
	cleanUpGateway = false
	return runningGatewayInfo, nil
}

// Calls `GetServices` and waits for the gateway to be ready
func waitForGatewayReady(apiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo) error {
	backgroundCtx := context.Background()
	gatewayAddress := fmt.Sprintf("%v:%v", apiContainerHostMachineInfo.IpOnHostMachine, apiContainerHostMachineInfo.GrpcPortOnHostMachine)

	conn, err := grpc.Dial(gatewayAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be dial in to API container running at address '%v', instead a non-nil error was returned", gatewayAddress)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	// The GRPC Server for our `API Container Gateway` is spun up in a gofunc
	// We call `GetServices` with `WaitForReady` to wait for the server to finish setting up
	// Modifying
	ctxWithTimeout, cancelFunc := context.WithTimeout(backgroundCtx, apiContainerGatewayHealthcheckTimeout)
	defer cancelFunc()
	getServicesHealthCheckParams := &kurtosis_core_rpc_api_bindings.GetServicesArgs{ServiceIdentifiers: nil}
	_, err = apiContainerClient.GetServices(ctxWithTimeout, getServicesHealthCheckParams, grpc.WaitForReady(waitForGatewayGrpcReady))
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be to call `GetServices` and wait for server to be ready, instead a non-nil error was returned")
	}

	return nil
}

func (service *EngineGatewayServiceServer) idempotentKillRunningGatewayForEnclaveId(enclaveId string) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	runningGateway, isRunning := service.enclaveIdToRunningGatewayMap[enclaveId]
	// Nothing running, nothing to kill
	if !isRunning {
		return
	}

	// Close up the connections
	runningGateway.closeFunc()

	logrus.Infof("Stopped running Gateway for enclave '%v' on port '%v'", enclaveId, runningGateway.hostMachineInfo.GrpcPortOnHostMachine)
	// delete the entry for the enclave
	delete(service.enclaveIdToRunningGatewayMap, enclaveId)
}
