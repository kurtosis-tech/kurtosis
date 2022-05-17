package engine_gateway

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/server/engine_gateway"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
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

func RunEngineGatewayUntilInterrupted(engine *engine.Engine, connectionProvider *connection.GatewayConnectionProvider) error {
	engineConnection, err := connectionProvider.ForEngine(engine)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to forward a local port to GRPC port of engine '%v', instead a non-nil error was returned", engine.GetID())
	}
	defer engineConnection.Stop()

	// Dial in to our locally forwarded port
	// remoteEngineClient -> kube-utils port-forwarded port (local) -> engine in cluster (remote)
	remoteEngineConn, err := engineConnection.GetGrpcClientConn()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get a GRPC client connection to the engine through a locally forwarded port, instead a non-nil error was returned")
	}
	defer remoteEngineConn.Close()
	engineClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(remoteEngineConn)

	engineGatewayServer, gatewayCloseFunc := engine_gateway.NewEngineGatewayServiceServer(connectionProvider, engineClient)
	defer gatewayCloseFunc()
	engineGatewayServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_engine_rpc_api_bindings.RegisterEngineServiceServer(grpcServer, engineGatewayServer)
	}
	// Print information to the user
	logrus.Infof("Starting the gateway for engine with ID '%v' on local port '%v'", engine.GetID(), engineGatewayPort)
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
