package api_container_gateway

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"sync"
)

type ApiContainerGatewayServiceServer struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	// Client for the api container we'll be connecting too
	remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	// Provides connections to Kurtosis objectis in cluster
	connectionProvider *connection.GatewayConnectionProvider

	// ServiceMap
	mutex                           *sync.Mutex
	userServiceToLocalConnectionMap map[string]connection.GatewayConnectionToKurtosis
}

func NewEnclaveApiContainerGatewayServer(connectionProvider *connection.GatewayConnectionProvider, remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient) (resultCoreGatewayServerService *ApiContainerGatewayServiceServer, resultGatewayCloseFunc func()) {
	// Start out with 0 connections to user services
	userServiceToLocalConnectionMap := map[string]connection.GatewayConnectionToKurtosis{}

	closeGatewayFunc := func() {
		// Stop any port forwarding
		for _, conn := range resultCoreGatewayServerService.userServiceToLocalConnectionMap {
			conn.Stop()
		}
	}

	return &ApiContainerGatewayServiceServer{
		remoteApiContainerClient:        remoteApiContainerClient,
		connectionProvider:              connectionProvider,
		mutex:                           &sync.Mutex{},
		userServiceToLocalConnectionMap: userServiceToLocalConnectionMap,
	}, closeGatewayFunc
}
