package main

import (
	"context"
	"github.com/bufbuild/connect-go"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings/kurtosis_engine_rpc_api_bindingsconnect"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"time"
)

const (
	listenPort                = 8081
	grpcServerStopGracePeriod = 5 * time.Second
)

type WebServer struct {
}

func NewWebserver() *WebServer {
	return &WebServer{}
}

func (c *WebServer) Check(
	context.Context,
	*connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest],
) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	response := &connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse]{
		Msg: &kurtosis_enclave_manager_api_bindings.HealthCheckResponse{
			Status: 1,
		},
	}
	return response, nil
}
func (c *WebServer) GetEnclaves(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	client := kurtosis_engine_rpc_api_bindingsconnect.NewEngineServiceClient(
		http.DefaultClient,
		"http://localhost:9710",
	)
	enclaves, err := client.GetEnclaves(ctx, req)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse]{
		Msg: &kurtosis_engine_rpc_api_bindings.GetEnclavesResponse{
			EnclaveInfo: enclaves.Msg.EnclaveInfo,
		},
	}
	return resp, nil
}
func (c *WebServer) GetServices(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}

func (c *WebServer) CreateEnclave(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}
func (c *WebServer) GetServiceLogs(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}
func (c *WebServer) RunStarlarkPackage(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, nil
}
func (c *WebServer) ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
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
