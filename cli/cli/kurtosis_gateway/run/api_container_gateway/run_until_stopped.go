package api_container_gateway

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/server/api_container_gateway"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"time"
)

const (
	grpcServerStopGracePeriod = 5 * time.Second
)

func RunApiContainerGatewayUntilStopped(connectionProvider *connection.GatewayConnectionProvider, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, gatewayPort uint16, gatewayStopChannel chan interface{}) error {
	apiContainerConnection, err := connectionProvider.ForEnclaveApiContainer(enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to start forwarding ports to an enclave API container, instead a non nil error was returned")
	}
	defer apiContainerConnection.Stop()

	// Dial in to our locally forwarded port
	apiContainerGrpcClientConn, err := apiContainerConnection.GetGrpcClientConn()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to create a grpc client connection to the forwarded API container port, instead a non nil error was returned")
	}
	defer apiContainerGrpcClientConn.Close()

	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(apiContainerGrpcClientConn)
	apiContainerGatewayServer, gatewayCloseFunc := api_container_gateway.NewEnclaveApiContainerGatewayServer(connectionProvider, apiContainerClient, enclaveInfo.GetEnclaveId())
	defer gatewayCloseFunc()

	apiContainerGatewayServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_core_rpc_api_bindings.RegisterApiContainerServiceServer(grpcServer, apiContainerGatewayServer)
	}
	apiContainerGatewayGrpcServer := minimal_grpc_server.NewMinimalGRPCServer(
		gatewayPort,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			apiContainerGatewayServiceRegistrationFunc,
		},
	)

	if err := apiContainerGatewayGrpcServer.RunUntilStopped(gatewayStopChannel); err != nil {
		return stacktrace.Propagate(err, "Expected to run API container gateway server until stopped, but the server exited with a non-nil error")
	}

	return nil
}
