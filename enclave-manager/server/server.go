package main

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
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

func (c *WebServer) Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	response := &connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse]{
		Msg: &kurtosis_enclave_manager_api_bindings.HealthCheckResponse{
			Status: 1,
		},
	}
	return response, nil
}
func (c *WebServer) GetEnclaves(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
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
func (c *WebServer) GetServices(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the APIC client")
	}

	serviceRequest := &connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesArgs{
			ServiceIdentifiers: map[string]bool{},
		},
	}
	serviceInfoMapFromAPIC, err := (*apiContainerServiceClient).GetServices(ctx, serviceRequest)
	if err != nil {
		return nil, err
	}

	resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesResponse{
			ServiceInfo: serviceInfoMapFromAPIC.Msg.GetServiceInfo(),
		},
	}
	return resp, nil
}

func (c *WebServer) GetServiceLogs(
	ctx context.Context,
	req *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs],
	str *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse],
) error {

	result, err := (*c.engineServiceClient).GetServiceLogs(ctx, req)
	if err != nil {
		return err
	}

	logs := getServiceLogsFromEngine(result)
	for {
		select {
		case <-ctx.Done():
			err := result.Close()
			if err != nil {
				logrus.Errorf("Error ocurred: %+v", err)
			}
			close(logs)
			return nil
		case resp := <-logs:
			errWhileSending := str.Send(resp)
			if errWhileSending != nil {
				return errWhileSending
			}
		}
	}
}

func (c *WebServer) ListFilesArtifactNamesAndUuids(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the APIC client")
	}

	serviceRequest := &connect.Request[emptypb.Empty]{}
	result, err := (*apiContainerServiceClient).ListFilesArtifactNamesAndUuids(ctx, serviceRequest)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse{
			FileNamesAndUuids: result.Msg.FileNamesAndUuids,
		},
	}
	return resp, nil
}

func (c *WebServer) RunStarlarkPackage(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest], str *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	runPackageArgs := req.Msg.RunStarlarkPackageArgs
	boolean := true
	runPackageArgs.ClonePackage = &boolean

	if err != nil {
		return stacktrace.Propagate(err, "Failed to create the APIC client")
	}
	serviceRequest := &connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs]{
		Msg: req.Msg.RunStarlarkPackageArgs,
	}

	apicStream, err := (*apiContainerServiceClient).RunStarlarkPackage(ctx, serviceRequest)
	ctxWithCancel, cancel := context.WithCancel(ctx)

	logs := getRuntimeLogsWhenCreatingEnclave(cancel, apicStream)
	for {
		select {
		case <-ctxWithCancel.Done():
			logrus.Infof("Closing the stream")
			err := apicStream.Close()
			if err != nil {
				logrus.Errorf("Error ocurred: %+v", err)
			}
			close(logs)
			return nil
		case resp := <-logs:
			errWhileSending := str.Send(resp)
			if errWhileSending != nil {
				logrus.Errorf("error occurred: %+v", errWhileSending)
				return nil
			}
		}
	}
}

func (c *WebServer) CreateEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	result, err := (*c.engineServiceClient).CreateEnclave(ctx, req)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse]{
		Msg: &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
			EnclaveInfo: result.Msg.EnclaveInfo,
		},
	}
	logrus.Infof("Create Enclave: %+v", resp)
	return resp, nil
}

func (c *WebServer) InspectFilesArtifactContents(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the APIC client")
	}

	request := &connect.Request[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest]{
		Msg: &kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest{
			FileNamesAndUuid: req.Msg.FileNamesAndUuid,
		},
	}
	result, err := (*apiContainerServiceClient).InspectFilesArtifactContents(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse{
			FileDescriptions: result.Msg.FileDescriptions,
		},
	}
	return resp, nil
}

func (c *WebServer) createAPICClient(
	ip string,
	port int32,
) (*kurtosis_core_rpc_api_bindingsconnect.ApiContainerServiceClient, error) {
	host, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, port))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse the connection url for the APIC")
	}
	apiContainerServiceClient := kurtosis_core_rpc_api_bindingsconnect.NewApiContainerServiceClient(
		http.DefaultClient,
		host.String(),
		connect.WithGRPCWeb(),
	)
	return &apiContainerServiceClient, nil
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

func getServiceLogsFromEngine(client *connect.ServerStreamForClient[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]) chan *kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse {
	result := make(chan *kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse)
	go func() {
		for client.Receive() {
			res := client.Msg()
			result <- res
		}
	}()
	return result
}

func getRuntimeLogsWhenCreatingEnclave(cancel context.CancelFunc, client *connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	result := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		for client.Receive() {
			res := client.Msg()
			result <- res
		}
		cancel()
	}()
	return result
}
