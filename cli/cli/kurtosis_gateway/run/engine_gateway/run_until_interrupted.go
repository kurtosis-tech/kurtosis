package engine_gateway

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/live_engine_client_supplier"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/server/engine_gateway"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

const (
	localHostIpStr            = "127.0.0.1"
	engineGatewayPort         = 9710
	grpcServerStopGracePeriod = 5 * time.Second
)

func RunEngineGatewayUntilInterrupted(kurtosisBackend backend_interface.KurtosisBackend, connectionProvider *connection.GatewayConnectionProvider) error {
	engineClientSupplier := live_engine_client_supplier.NewLiveEngineClientSupplier(kurtosisBackend, connectionProvider)
	if err := engineClientSupplier.Start(); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to start supplier for live Kurtosis engine clients, instead a non-nil error was returned")
	}
	engineGatewayServer, gatewayCloseFunc := engine_gateway.NewEngineGatewayServiceServer(connectionProvider, engineClientSupplier)
	defer gatewayCloseFunc()
	engineGatewayServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_engine_rpc_api_bindings.RegisterEngineServiceServer(grpcServer, engineGatewayServer)
	}
	// Print information to the user
	logrus.Infof("Starting Kurtosis gateway on local port '%v'", engineGatewayPort)
	logrus.Infof("You can use this gateway as a drop-in replacement for Kurtosis engine. To connect to the gateway, send a request to '%v:%v'", localHostIpStr, engineGatewayPort)
	logrus.Infof("To kill the running gateway, press CTRL+C")

	engineGatewayGrpcServer := minimal_grpc_server.NewMinimalGRPCServer(
		engineGatewayPort,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			engineGatewayServiceRegistrationFunc,
		},
	)

	if err := engineGatewayGrpcServer.RunUntilInterrupted(); err != nil {
		return stacktrace.Propagate(err, "Expected to run Engine gateway server until interrupted, but the server exited with a non-nil error")
	}

	return nil

}
