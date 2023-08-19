package main

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-go"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings/kurtosis_engine_rpc_api_bindingsconnect"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
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
	request *connect.Request[emptypb.Empty],
) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	enclaves, err := (*c.engineServiceClient).GetEnclaves(ctx, request)
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
	// buf bindings don't seem to work as I expect:
	//host, err := url.Parse(fmt.Sprintf("http://%s:%d", ipAddress, port))
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "Failed to parse the connection url")
	//}
	//logrus.Infof("Calling APIC: %s", host.String())
	//apiContainerServiceClient := kurtosis_core_rpc_api_bindingsconnect.NewApiContainerServiceClient(
	//	http.DefaultClient,
	//	host.String(),
	//)
	//
	//serviceRequest := &connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]{
	//	Msg: &kurtosis_core_rpc_api_bindings.GetServicesArgs{
	//		ServiceIdentifiers: map[string]bool{},
	//	},
	//}
	//res, err := apiContainerServiceClient.GetServices(ctx, serviceRequest)
	host := fmt.Sprintf("%s:%d", ipAddress, port)
	logrus.Infof("Calling APIC: %s", host)

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container grpc port at '%v'",
			host,
		)
	}
	defer func() {
		conn.Close()
	}()

	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)
	getAllServicesMap := map[string]bool{}
	getAllServicesArgs := binding_constructors.NewGetServicesArgs(getAllServicesMap)
	allServicesResponse, err := apiContainerClient.GetServices(ctx, getAllServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service information for all services in APIC '%v'", host)
	}
	serviceInfoMapFromAPIC := allServicesResponse.GetServiceInfo()
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesResponse{
			ServiceInfo: serviceInfoMapFromAPIC,
		},
	}
	return resp, nil
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
