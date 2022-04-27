package engine_gateway

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/port_utils"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/run/api_container_gateway"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	restclient "k8s.io/client-go/rest"
	"sync"
	"time"
)

const (
	gatewayPortRangeStart     = 33000
	grpcServerStopGracePeriod = 5 * time.Second
	localHostIpStr            = "127.0.0.1"

	tcpProtocolStr              = "tcp"
	localHostZeroPortBindingStr = "localhost:0"
)

type EngineGatewayServiceServer struct {
	// This embedding is required by gRPC
	kurtosis_engine_rpc_api_bindings.UnimplementedEngineServiceServer

	// Client for the engine we'll be connecting too
	remoteEngineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient

	// Configuration for the kubernetes cluster kurtosis is running on
	kubernetesConfig *restclient.Config

	connectionProvider *connection.GatewayConnectionProvider

	mutex                        *sync.Mutex
	enclaveIdToRunningGatewayMap map[string]*runningApiContainerGateway

	// Used to assign ports for gateway GRPC servers to run on
	// Incremented each time an api container gateway server is starts
	portForEnclaveGateway uint16
}

type runningApiContainerGateway struct {
	// info about the api container on the host machine
	hostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	// closeGatewayFunc
	closeFunc func()
}

// NewEngineGatewayServiceServer returns a EngineGatewayServiceServer
// runningEngine is a kurtosis engine running a cluster that can be reached through clients configured with kubernetesConfig
func NewEngineGatewayServiceServer(connectionProvider *connection.GatewayConnectionProvider, engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient) (resultGatewayService *EngineGatewayServiceServer, gatewayCloseFunc func()) {
	// We start out with no enclave api-container gateways runnings
	runningApiContainers := map[string]*runningApiContainerGateway{}

	service := &EngineGatewayServiceServer{
		remoteEngineClient:           engineClient,
		connectionProvider:           connectionProvider,
		enclaveIdToRunningGatewayMap: runningApiContainers,
		mutex:                        &sync.Mutex{},
		// local port
		portForEnclaveGateway: gatewayPortRangeStart,
	}
	closeFunc := func() {
		// Kill the running enclave gateways
		for enclaveId, _ := range service.enclaveIdToRunningGatewayMap {
			service.idempotentKillRunningGatewayForEnclaveId(enclaveId)
		}
	}

	return service, closeFunc
}

func (service *EngineGatewayServiceServer) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error) {
	remoteEngineResponse, err := service.remoteEngineClient.CreateEnclave(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave through the remote engine")
	}
	cleanUpEnclave := true
	defer func() {
		if cleanUpEnclave {
			destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{EnclaveId: args.GetEnclaveId()}
			if _, err := service.remoteEngineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
				logrus.Error("Launching the Enclave gateway failed, expected to be able to cleanup the created enclave, but an error occurred calling the backend to destroy the enclave we created:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill the enclave with id '%v'", args.GetEnclaveId())
			}
		}
	}()
	// Update enclave API Container Host Machine Info
	// We want to update the host machine info for the api container
	createdEnclaveInfo := remoteEngineResponse.GetEnclaveInfo()
	createdEnclaveId := createdEnclaveInfo.GetEnclaveId()

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
	remoteEngineResponse, err := service.remoteEngineClient.GetEnclaves(ctx, in)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves from the remote engine")
	}
	responseEnclaves := remoteEngineResponse.GetEnclaveInfo()
	for enclaveId, enclaveInfo := range responseEnclaves {
		var runningApiContainerGateway *runningApiContainerGateway
		runningApiContainerGateway, isRunning := service.enclaveIdToRunningGatewayMap[enclaveId]
		// If the gateway isn't running, start it
		if !isRunning {
			runningApiContainerGateway, err = service.startRunningGatewayForEnclave(enclaveInfo)
		}
		remoteEngineResponse.EnclaveInfo[enclaveId].ApiContainerHostMachineInfo = runningApiContainerGateway.hostMachineInfo
	}
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) StopEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.StopEnclaveArgs) (*emptypb.Empty, error) {
	if _, err := service.remoteEngineClient.StopEnclave(ctx, args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred calling remote engine to stop enclave '%v'", args.EnclaveId)
	}

	return &emptypb.Empty{}, nil
}

func (service *EngineGatewayServiceServer) DestroyEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs) (*emptypb.Empty, error) {
	if _, err := service.remoteEngineClient.DestroyEnclave(ctx, args); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred calling remote engine to destroy enclave with ID '%v':", args.EnclaveId)
	}
	// Kill the running api container gateway for this enclave
	enclaveIdOfGatewayToKill := args.EnclaveId
	service.idempotentKillRunningGatewayForEnclaveId(enclaveIdOfGatewayToKill)

	return &emptypb.Empty{}, nil
}

func (service *EngineGatewayServiceServer) Clean(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CleanArgs) (*kurtosis_engine_rpc_api_bindings.CleanResponse, error) {
	remoteEngineResponse, err := service.remoteEngineClient.Clean(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}
	gatewaysToKill := remoteEngineResponse.RemovedEnclaveIds
	for gatewayEnclaveId := range gatewaysToKill {
		service.idempotentKillRunningGatewayForEnclaveId(gatewayEnclaveId)
	}
	return remoteEngineResponse, nil
}

func (service *EngineGatewayServiceServer) GetEngineInfo(ctx context.Context, emptyArgs *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	remoteEngineResponse, err := service.remoteEngineClient.GetEngineInfo(ctx, emptyArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engine info through the remote engine")
	}
	return remoteEngineResponse, nil
}

// Private functions for managing our running enclave api container gateways
func (service *EngineGatewayServiceServer) startRunningGatewayForEnclave(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) (*runningApiContainerGateway, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	var isGatewayRunning = false
	enclaveId := enclaveInfo.GetEnclaveId()
	// Ask the kernel for a free open TCP port
	gatewayPortSpec, err := port_utils.GetFreeTcpPort()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get a free, open TCP port on host, instead a non-nil error was returned")
	}
	// Channel for messages to stop the running server
	gatewayServerStopper := make(chan interface{}, 1)
	// Stop the running gateway
	gatewayStopFunc := func() {
		// Send message to stop the gateway server
		gatewayServerStopper <- nil
	}
	defer gatewayStopFunc()
	// Info for how to connect to the api container through the gateway running on host machine
	apiContainerHostMachineInfo := &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
		IpOnHostMachine:       localHostIpStr,
		GrpcPortOnHostMachine: uint32(gatewayPortSpec.GetNumber()),
		// TODO proxy endpoint for gateway
		GrpcProxyPortOnHostMachine: 0,
	}

	// Start the server in a goroutine
	go func() {
		defer gatewayStopFunc()
		if err := api_container_gateway.RunApiContainerGatewayUntilStopped(service.connectionProvider, enclaveInfo, gatewayServerStopper); err != nil {
			logrus.Warnf("Expected to run api container gateway until stopped, but the server exited prematurely with a non-nil error: '%v'", err)
		}
	}()

	runningGatewayInfo := &runningApiContainerGateway{
		closeFunc:       gatewayStopFunc,
		hostMachineInfo: apiContainerHostMachineInfo,
	}

	// Store information about our running gateway
	service.enclaveIdToRunningGatewayMap[enclaveId] = runningGatewayInfo
	defer func() {
		if !isGatewayRunning {
			delete(service.enclaveIdToRunningGatewayMap, enclaveId)
		}
	}()
	isGatewayRunning = true
	// Increment the port number for the next gateway server we start
	// Ideally, we would rely on the host to assign us a port, but minimal_grpc_server does not support this atm
	service.portForEnclaveGateway = service.portForEnclaveGateway + 1
	return runningGatewayInfo, nil
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
	// delete the entry for the enclave
	delete(service.enclaveIdToRunningGatewayMap, enclaveId)
}
