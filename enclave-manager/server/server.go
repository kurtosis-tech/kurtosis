package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings/kurtosis_core_rpc_api_bindingsconnect"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings/kurtosis_engine_rpc_api_bindingsconnect"
	"github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings/kurtosis_backend_server_rpc_api_bindingsconnect"
	connect_server "github.com/kurtosis-tech/kurtosis/connect-server"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	listenPort                                       = 8081
	grpcServerStopGracePeriod                        = 5 * time.Second
	engineHostUrl                                    = "http://localhost:9710"
	kurtosisCloudApiHost                             = "https://cloud.kurtosis.com"
	kurtosisCloudApiPort                             = 8080
	KurtosisEnclaveManagerApiEnforceAuthKeyEnvVarArg = "KURTOSIS_ENCLAVE_MANAGER_API_ENFORCE_AUTH"
)

type Authentication struct {
	ApiKey   string
	JwtToken string
}

type WebServer struct {
	engineServiceClient *kurtosis_engine_rpc_api_bindingsconnect.EngineServiceClient
	enforceAuth         bool
}

func NewWebserver(enforceAuth bool) (*WebServer, error) {
	engineServiceClient := kurtosis_engine_rpc_api_bindingsconnect.NewEngineServiceClient(
		http.DefaultClient,
		engineHostUrl,
	)
	return &WebServer{
		engineServiceClient: &engineServiceClient,
		enforceAuth:         enforceAuth,
	}, nil
}

func (c *WebServer) Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	response := &connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse]{
		Msg: &kurtosis_enclave_manager_api_bindings.HealthCheckResponse{
			Status: 1,
		},
	}
	return response, nil
}

func (c *WebServer) ValidateRequestAuthorization(
	ctx context.Context,
	enforceAuth bool,
	header http.Header,
) (bool, error) {
	if !enforceAuth {
		return true, nil
	}

	reqToken := header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		return false, stacktrace.NewError("Authorization token malformed. Bearer token format required")
	}
	reqToken = strings.TrimSpace(splitToken[1])
	auth, err := c.ConvertJwtTokenToApiKey(ctx, reqToken)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to convert jwt token to API key")
	}
	if auth == nil || len(auth.ApiKey) == 0 {
		return false, stacktrace.NewError("An internal error has occurred. An empty API key was found")
	}

	instanceConfig, err := c.GetCloudInstanceConfig(ctx, auth.ApiKey)
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to retrieve the instance config")
	}
	reqHost := header.Get("Host")
	splitHost := strings.Split(reqHost, ":")
	if len(splitHost) != 2 {
		return false, stacktrace.NewError("Host header malformed. host:port format required")
	}
	reqHost = splitHost[0]
	if instanceConfig.LaunchResult.PublicDns != reqHost {
		return false, stacktrace.NewError("Instance config public dns '%s' does not match the request host '%s'", instanceConfig.LaunchResult.PublicDns, reqHost)
	}

	return true, nil
}

func (c *WebServer) GetEnclaves(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	auth, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
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
	auth, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
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
	auth, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
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
	logrus.Infof("YOLO: %+v", req)

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

	logrus.Debugf("Hellooo %+v", serviceRequest)
	logs := getRuntimeLogsWhenCreatingEnclave(cancel, apicStream)
	for {
		select {
		case <-ctxWithCancel.Done():
			err := apicStream.Close()
			close(logs)
			if err != nil {
				return stacktrace.Propagate(err, "Error occurred after closing the stream")
			}
			return nil
		case resp := <-logs:
			errWhileSending := str.Send(resp)
			if errWhileSending != nil {
				logrus.Errorf("error occurred: %+v", errWhileSending)
				return stacktrace.Propagate(errWhileSending, "Error occurred while sending streams")
			}
		}
	}
}

func (c *WebServer) CreateEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	auth, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
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
	auth, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
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

func (c *WebServer) createKurtosisCloudBackendClient(
	host string,
	port int,
) (*kurtosis_backend_server_rpc_api_bindingsconnect.KurtosisCloudBackendServerClient, error) {
	url, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse the connection url for Kurtosis Cloud Backend")
	}
	client := kurtosis_backend_server_rpc_api_bindingsconnect.NewKurtosisCloudBackendServerClient(
		http.DefaultClient,
		url.String(),
		connect.WithGRPCWeb(),
	)
	return &client, nil
}

func (c *WebServer) GetCloudInstanceConfig(
	ctx context.Context,
	apiKey string,
) (*kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse, error) {
	client, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}
	getInstanceRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest{
			ApiKey: apiKey,
		},
	}
	getInstanceResponse, err := (*client).GetOrCreateInstance(ctx, getInstanceRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get the instance")
	}
	getInstanceConfigRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs{
			ApiKey:     apiKey,
			InstanceId: getInstanceResponse.Msg.InstanceId,
		},
	}
	getInstanceConfigResponse, err := (*client).GetCloudInstanceConfig(ctx, getInstanceConfigRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get the instance config")
	}

	return getInstanceConfigResponse.Msg, nil
}

func (c *WebServer) ConvertJwtTokenToApiKey(
	ctx context.Context,
	jwtToken string,
) (*Authentication, error) {
	client, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}
	request := &connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest{
			AccessToken: jwtToken,
		},
	}
	result, err := (*client).GetOrCreateApiKey(ctx, request)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get the API key")
	}

	if result == nil {
		// User does not have an API key (unlikely if valid jwt token)
		return nil, stacktrace.NewError("User does not have an API key assigned")
	}
	if len(result.Msg.ApiKey) > 0 {
		return &Authentication{
			ApiKey:   result.Msg.ApiKey,
			JwtToken: jwtToken,
		}, nil
	}

	return nil, stacktrace.NewError("an empty API key was returned from Kurtosis Cloud Backend")
}

func RunEnclaveManagerApiServer(enforceAuth bool) error {
	srv, err := NewWebserver(enforceAuth)
	if err != nil {
		logrus.Fatal("an error occurred while processing the auth settings, exiting!", err)
		return err
	}
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
	return nil
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
