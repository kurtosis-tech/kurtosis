package main

import (
	"context"
	"github.com/bufbuild/connect-go"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	listenPort                = 8080
	grpcServerStopGracePeriod = 5 * time.Second
)

type WebServer struct {
}

func NewWebserver() *WebServer {
	return &WebServer{}
}

func (c *WebServer) Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	return nil, nil
}
func (c *WebServer) GetEnclaves(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}
func (c *WebServer) GetServices(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}

func RunEnclaveApiServer() {

	srv := NewWebserver()
	apiPath, handler := kurtosis_enclave_manager_api_bindingsconnect.NewKurtosisEnclaveManagerServerHandler(srv)

	logrus.Infof("Web server running and listening on port %d", listenPort)
	apiServer := connect_server.NewConnectServer(
		listenPort,
		grpcServerStopGracePeriod,
		handler,
		apiPath,
	)
	if err := apiServer.RunServerUntilInterruptedWithCors(cors.AllowAll()); err != nil {
		logrus.Error("An error occurred running the server", err)
	}

}
