package main

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings/kurtosis_core_rpc_api_bindingsconnect"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings/kurtosis_engine_rpc_api_bindingsconnect"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"net/url"
	"time"
)

const (
	listenPort                = 8081
	grpcServerStopGracePeriod = 5 * time.Second
	engineHostUrl             = "http://localhost:9710"
)

type WebServer struct {
	engineServiceClient *kurtosis_engine_rpc_api_bindingsconnect.EngineServiceClient
}

func NewWebserver() *WebServer {
	engineServiceClient := kurtosis_engine_rpc_api_bindingsconnect.NewEngineServiceClient(
		http.DefaultClient,
		engineHostUrl,
	)
	return &WebServer{
		engineServiceClient: &engineServiceClient,
	}
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
	enclaves, err := (*c.engineServiceClient).GetEnclaves(ctx, req)
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
func (c *WebServer) GetServices(
	ctx context.Context,
	request *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {

	ipAddress := request.Msg.ApicIpAddress
	port := request.Msg.ApicPort

	host, err := url.Parse(fmt.Sprintf("http://%s:%d", ipAddress, port))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse the connection url")
	}
	logrus.Infof("Calling APIC: %s", host.String())
	apiContainerServiceClient := kurtosis_core_rpc_api_bindingsconnect.NewApiContainerServiceClient(
		http.DefaultClient,
		host.String(),
		connect.WithGRPCWeb(),
	)

	serviceRequest := &connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesArgs{
			ServiceIdentifiers: map[string]bool{},
		},
	}
	serviceInfoMapFromAPIC, err := apiContainerServiceClient.GetServices(ctx, serviceRequest)

	resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesResponse{
			ServiceInfo: serviceInfoMapFromAPIC.Msg.GetServiceInfo(),
		},
	}
	return resp, nil
}

func (c *WebServer) CreateEnclave(
	ctx context.Context,
	req *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs],
) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	result, err := (*c.engineServiceClient).CreateEnclave(ctx, req)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse]{
		Msg: &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
			EnclaveInfo: result.Msg.EnclaveInfo,
		},
	}
	return resp, nil
}
func (c *WebServer) GetServiceLogs(
	ctx context.Context,
	req *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs],
) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse], error) {
	result, err := (*c.engineServiceClient).GetServiceLogs(ctx, req)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]{
		Msg: &kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse{
			ServiceLogsByServiceUuid: result.Msg().ServiceLogsByServiceUuid,
			NotFoundServiceUuidSet:   result.Msg().NotFoundServiceUuidSet,
		},
	}
	return resp, nil
}
func (c *WebServer) RunStarlarkPackage(
	context.Context,
	*connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest],
) (*connect.Response[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error) {
	return nil, nil
}
func (c *WebServer) ListFilesArtifactNamesAndUuids(
	context.Context,
	*connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest],
) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
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
