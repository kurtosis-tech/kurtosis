package server

import (
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/engine_functions/github_auth_storage_creator"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
	listenPort                 = 8081
	grpcServerStopGracePeriod  = 5 * time.Second
	engineHostUrl              = "http://localhost:9710"
	kurtosisCloudApiHost       = "https://cloud.kurtosis.com"
	kurtosisCloudApiPort       = 8080
	numberOfElementsAuthHeader = 2
	numberOfElementsHostString = 2
	slashSeparator             = "/"
	shortUuidLength            = 12
)

type Authentication struct {
	ApiKey   string
	JwtToken string
}

type WebServer struct {
	instanceConfigMutex *sync.RWMutex
	apiKeyMutex         *sync.RWMutex
	engineServiceClient *kurtosis_engine_rpc_api_bindingsconnect.EngineServiceClient
	enforceAuth         bool
	instanceConfig      *kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse
	instanceConfigMap   map[string]*kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse
	apiKeyMap           map[string]*string
	githubAccessToken   string
}

func NewWebserver(enforceAuth bool) (*WebServer, error) {
	engineServiceClient := kurtosis_engine_rpc_api_bindingsconnect.NewEngineServiceClient(
		http.DefaultClient,
		engineHostUrl,
	)
	githubAuthToken := github_auth_storage_creator.GetGitHubAuthToken()
	return &WebServer{
		engineServiceClient: &engineServiceClient,
		enforceAuth:         enforceAuth,
		instanceConfigMutex: &sync.RWMutex{},
		apiKeyMutex:         &sync.RWMutex{},
		apiKeyMap:           map[string]*string{},
		instanceConfigMap:   map[string]*kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse{},
		instanceConfig:      nil,
		githubAccessToken:   githubAuthToken,
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

func (c *WebServer) CreateRepositoryWebhook(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.CreateRepositoryWebhookRequest]) (*connect.Response[emptypb.Empty], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}

	if c.githubAccessToken == "" {
		return nil, stacktrace.NewError("GitHub AuthToken is empty for this enclave manager. This method shouldn't be called")
	}
	packageId := req.Msg.PackageId
	packageIdSplit := strings.Split(packageId, slashSeparator)
	owner := packageIdSplit[len(packageIdSplit)-2]
	repo := packageIdSplit[len(packageIdSplit)-1]
	client := github.NewClient(nil).WithAuthToken(c.githubAccessToken)
	webhookUrl := "https://preview.kurtosis.com/webhook"
	contentTypeJson := "json"
	hook := &github.Hook{
		Name: github.String("web"),
		Config: &github.HookConfig{
			URL:         &webhookUrl,
			ContentType: &contentTypeJson,
			InsecureSSL: nil,
			Secret:      nil,
		},
		Events:       []string{"push", "pull_request"},
		Active:       github.Bool(true),
		CreatedAt:    nil,
		UpdatedAt:    nil,
		URL:          nil,
		ID:           nil,
		Type:         nil,
		TestURL:      nil,
		PingURL:      nil,
		LastResponse: nil,
	}
	_, _, err = client.Repositories.CreateHook(ctx, owner, repo, hook)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while creating the webhook")
	}
	return &connect.Response[emptypb.Empty]{}, nil
}

func (c *WebServer) ValidateRequestAuthorization(
	ctx context.Context,
	enforceAuth bool,
	header http.Header,
) (bool, *kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse, error) {
	if !enforceAuth {
		return true, nil, nil
	}
	reqToken, err := extractJwtToken(header)
	if err != nil {
		return false, nil, err
	}
	auth, err := c.ConvertJwtTokenToApiKey(ctx, reqToken)
	if err != nil {
		return false, nil, stacktrace.Propagate(err, "Failed to convert jwt token to API key")
	}
	if auth == nil || len(auth.ApiKey) == 0 {
		return false, nil, stacktrace.NewError("An internal error has occurred. An empty API key was found")
	}

	instanceConfig, err := c.getCloudInstanceConfig(ctx, reqToken, auth.ApiKey)
	if err != nil {
		return false, nil, stacktrace.Propagate(err, "Failed to retrieve the instance config")
	}
	reqHost := header.Get("Host")
	splitHost := strings.Split(reqHost, ":")
	if len(splitHost) != numberOfElementsHostString {
		return false, nil, stacktrace.NewError("Host header malformed. host:port format required")
	}
	reqHost = splitHost[0]
	if instanceConfig.LaunchResult.PublicDns != reqHost {
		delete(c.apiKeyMap, reqToken)
		delete(c.instanceConfigMap, reqToken)
		return false, nil, stacktrace.NewError("either the requested instance does not exist or the user is not authorized to access the resource")
	}

	return true, instanceConfig, nil
}

func (c *WebServer) GetCloudInstanceConfig(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error) {
	if !c.enforceAuth {
		return nil, stacktrace.NewError("This method is only available in the cloud")
	}

	jwtToken, err := extractJwtToken(req.Header())
	if err != nil {
		return nil, err
	}
	auth, err := c.ConvertJwtTokenToApiKey(ctx, jwtToken)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert jwt token to API key")
	}
	if auth == nil || len(auth.ApiKey) == 0 {
		return nil, stacktrace.NewError("An internal error has occurred. An empty API key was found")
	}

	instanceConfig, err := c.getCloudInstanceConfig(ctx, jwtToken, auth.ApiKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to retrieve the instance config")
	}

	resp := &connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse]{
		Msg: instanceConfig,
	}

	return resp, nil
}

func (c *WebServer) GetEnclaves(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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
	auth, instanceConfig, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

	serviceInfoMapFromApicObj := serviceInfoMapFromAPIC.Msg.GetServiceInfo()

	// we aren't in a cloud context, so we exit early
	if !c.enforceAuth {
		resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse]{
			Msg: &kurtosis_core_rpc_api_bindings.GetServicesResponse{
				ServiceInfo: serviceInfoMapFromApicObj,
			},
		}
		return resp, nil
	}

	cloudClient, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}

	jwtToken, err := extractJwtToken(req.Header())
	if err != nil {
		return nil, err
	}
	getUnlockedPortsRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.GetPortsRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.GetPortsRequest{
			AccessToken:       jwtToken,
			InstanceShortUuid: instanceConfig.InstanceId[:shortUuidLength],
			EnclaveShortUuid:  req.Msg.EnclaveShortenedUuid,
		},
	}

	var portMetadata []*kurtosis_backend_server_rpc_api_bindings.CloudPort
	cloudPortsResponse, err := (*cloudClient).GetPorts(ctx, getUnlockedPortsRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while pulling unauthenticated ports from the cloud backend")
	}
	portMetadata = cloudPortsResponse.Msg.CloudPorts

	for _, service := range serviceInfoMapFromApicObj {
		serviceShortUuid := service.ServiceUuid[:shortUuidLength]
		for _, privatePort := range service.PrivatePorts {
			locked := true
			alias := ""
			for _, unlockedPortsAndService := range portMetadata {
				if unlockedPortsAndService.Port.ServiceShortUuid == serviceShortUuid {
					if privatePort.Number == unlockedPortsAndService.Port.PortNumber {
						locked = unlockedPortsAndService.Locked
						alias = unlockedPortsAndService.Alias
						break
					}
				}
			}
			privatePort.Locked = &locked
			privatePort.Alias = &alias
		}
	}

	resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.GetServicesResponse{
			ServiceInfo: serviceInfoMapFromApicObj,
		},
	}

	return resp, nil
}

func (c *WebServer) AddAlias(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.AddAliasRequest]) (*connect.Response[emptypb.Empty], error) {
	if !c.enforceAuth {
		return nil, stacktrace.NewError("This method is only available in the cloud")
	}
	auth, instanceConfig, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}

	cloudClient, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}

	jwtToken, err := extractJwtToken(req.Header())
	if err != nil {
		return nil, err
	}
	addAliasRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.AddAliasRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.AddAliasRequest{
			AccessToken: jwtToken,
			Port: &kurtosis_backend_server_rpc_api_bindings.Port{
				InstanceShortUuid: instanceConfig.InstanceId[:shortUuidLength],
				PortNumber:        req.Msg.PortNumber,
				EnclaveShortUuid:  req.Msg.EnclaveShortUuid,
				ServiceShortUuid:  req.Msg.ServiceShortUuid,
			},
			Alias: req.Msg.Alias,
		},
	}

	_, err = (*cloudClient).AddAlias(ctx, addAliasRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while sending add alias request to the cloud backend")
	}

	return &connect.Response[emptypb.Empty]{}, nil

}

func (c *WebServer) LockPort(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.LockUnlockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	if !c.enforceAuth {
		return nil, stacktrace.NewError("This method is only available in the cloud")
	}
	auth, instanceConfig, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}

	cloudClient, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}

	jwtToken, err := extractJwtToken(req.Header())
	if err != nil {
		return nil, err
	}
	lockPortRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.LockPortRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.LockPortRequest{
			AccessToken: jwtToken,
			Port: &kurtosis_backend_server_rpc_api_bindings.Port{
				InstanceShortUuid: instanceConfig.InstanceId[:shortUuidLength],
				PortNumber:        req.Msg.PortNumber,
				EnclaveShortUuid:  req.Msg.EnclaveShortUuid,
				ServiceShortUuid:  req.Msg.ServiceShortUuid,
			},
		},
	}

	_, err = (*cloudClient).LockPort(ctx, lockPortRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while sending lock port request to the cloud backend")
	}

	return &connect.Response[emptypb.Empty]{}, nil
}

func (c *WebServer) UnlockPort(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.LockUnlockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	if !c.enforceAuth {
		return nil, stacktrace.NewError("This method is only available in the cloud")
	}
	auth, instanceConfig, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}

	cloudClient, err := c.createKurtosisCloudBackendClient(
		kurtosisCloudApiHost,
		kurtosisCloudApiPort,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the Cloud backend client")
	}

	jwtToken, err := extractJwtToken(req.Header())
	if err != nil {
		return nil, err
	}
	lockPortRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest{
			AccessToken: jwtToken,
			Port: &kurtosis_backend_server_rpc_api_bindings.Port{
				InstanceShortUuid: instanceConfig.InstanceId[:shortUuidLength],
				PortNumber:        req.Msg.PortNumber,
				EnclaveShortUuid:  req.Msg.EnclaveShortUuid,
				ServiceShortUuid:  req.Msg.ServiceShortUuid,
			},
		},
	}

	_, err = (*cloudClient).UnlockPort(ctx, lockPortRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while sending unlock port request to the cloud backend")
	}

	return &connect.Response[emptypb.Empty]{}, nil
}

func (c *WebServer) GetServiceLogs(
	ctx context.Context,
	req *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs],
	str *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse],
) error {
	serviceLogsStream, err := (*c.engineServiceClient).GetServiceLogs(ctx, req)
	if err != nil {
		return err
	}

	for serviceLogsStream.Receive() {
		resp := serviceLogsStream.Msg()
		errWhileSending := str.Send(resp)
		if errWhileSending != nil {
			return stacktrace.Propagate(errWhileSending, "An error occurred in the enclave manager server attempting to send services logs.")
		}
	}
	if serviceLogsStream.Err() != nil {
		return stacktrace.Propagate(serviceLogsStream.Err(), "An error occurred in the enclave manager server attempting to receive services logs.")
	}

	return nil
}

func (c *WebServer) ListFilesArtifactNamesAndUuids(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

	// nolint: exhaustruct
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

func (c *WebServer) RunStarlarkPackage(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest], responseStream *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create the APIC client")
	}
	runStarlarkRequest := &connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs]{
		Msg: req.Msg.RunStarlarkPackageArgs,
	}

	runPackageArgs := req.Msg.RunStarlarkPackageArgs
	shouldClonePackage := true // ktoday: Why do we coerce the "clone" to true?
	runPackageArgs.ClonePackage = &shouldClonePackage

	starlarkLogsStream, err := (*apiContainerServiceClient).RunStarlarkPackage(ctx, runStarlarkRequest)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run package: %s", req.Msg.RunStarlarkPackageArgs.PackageId)
	}

	for starlarkLogsStream.Receive() {
		resp := starlarkLogsStream.Msg()
		err = responseStream.Send(resp)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to return logs from running the Starlark package.")
		}
	}
	if err = starlarkLogsStream.Err(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to return logs from running the Starlark package.")
	}

	return nil
}

func (c *WebServer) RunStarlarkScript(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkScriptRequest], responseStream *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create the APIC client")
	}

	runScriptArgs := req.Msg.RunStarlarkScriptArgs
	runStarlarkRequest := &connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs]{
		Msg: req.Msg.RunStarlarkScriptArgs,
	}

	starlarkLogsStream, err := (*apiContainerServiceClient).RunStarlarkScript(ctx, runStarlarkRequest)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run the following Starlark script:\n%s", runScriptArgs.SerializedScript)
	}

	for starlarkLogsStream.Receive() {
		resp := starlarkLogsStream.Msg()
		err = responseStream.Send(resp)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to return logs from running the Starlark script.")
		}
	}
	if err = starlarkLogsStream.Err(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to return logs from running the Starlark script.")
	}

	return nil
}

func (c *WebServer) DestroyEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return nil, stacktrace.Propagate(err, "User not authorized")
	}
	_, err = (*c.engineServiceClient).DestroyEnclave(ctx, req)
	if err != nil {
		return nil, err
	}
	// nolint: exhaustruct
	return &connect.Response[emptypb.Empty]{}, nil

}

func (c *WebServer) CreateEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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
	return resp, nil
}

func (c *WebServer) InspectFilesArtifactContents(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

func (c *WebServer) DownloadFilesArtifact(
	ctx context.Context,
	req *connect.Request[kurtosis_enclave_manager_api_bindings.DownloadFilesArtifactRequest],
	str *connect.ServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk],
) error {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
	if err != nil {
		return stacktrace.Propagate(err, "Authentication attempt failed")
	}
	if !auth {
		return stacktrace.Propagate(err, "User not authorized")
	}
	apiContainerServiceClient, err := c.createAPICClient(req.Msg.ApicIpAddress, req.Msg.ApicPort)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create the APIC client")
	}

	filesArtifactIdentifier := req.Msg.DownloadFilesArtifactsArgs.Identifier
	downloadFilesArtifactRequest := &connect.Request[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs]{
		Msg: &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs{
			Identifier: filesArtifactIdentifier,
		},
	}

	filesArtifactStream, err := (*apiContainerServiceClient).DownloadFilesArtifact(ctx, downloadFilesArtifactRequest)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create download stream for file artifact: %s", filesArtifactIdentifier)
	}
	for filesArtifactStream.Receive() {
		resp := filesArtifactStream.Msg()
		err = str.Send(resp)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to send streamed data chunks for files artifact with identifier: %v", filesArtifactIdentifier)
		}
	}
	if err = filesArtifactStream.Err(); err != nil {
		return stacktrace.Propagate(err, "An error occurred in the enclave manager server attempting to receive streamed data chunks for files artifact with identifier %v", filesArtifactIdentifier)
	}
	return nil
}

func (c *WebServer) GetStarlarkRun(
	ctx context.Context,
	req *connect.Request[kurtosis_enclave_manager_api_bindings.GetStarlarkRunRequest],
) (*connect.Response[kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

	request := &connect.Request[emptypb.Empty]{
		Msg: &emptypb.Empty{},
	}
	result, err := (*apiContainerServiceClient).GetStarlarkRun(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse]{
		Msg: &kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse{
			PackageId:              result.Msg.PackageId,
			SerializedScript:       result.Msg.SerializedScript,
			SerializedParams:       result.Msg.SerializedParams,
			Parallelism:            result.Msg.Parallelism,
			RelativePathToMainFile: result.Msg.RelativePathToMainFile,
			MainFunctionName:       result.Msg.MainFunctionName,
			ExperimentalFeatures:   result.Msg.ExperimentalFeatures,
			RestartPolicy:          result.Msg.RestartPolicy,
		},
	}
	return resp, nil
}

func (c *WebServer) GetStarlarkScriptPlanYaml(
	ctx context.Context,
	req *connect.Request[kurtosis_enclave_manager_api_bindings.StarlarkScriptPlanYamlArgs],
) (*connect.Response[kurtosis_core_rpc_api_bindings.PlanYaml], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

	request := &connect.Request[kurtosis_core_rpc_api_bindings.StarlarkScriptPlanYamlArgs]{
		Msg: &kurtosis_core_rpc_api_bindings.StarlarkScriptPlanYamlArgs{
			SerializedScript: req.Msg.StarlarkScriptPlanYamlArgs.SerializedScript,
			SerializedParams: req.Msg.StarlarkScriptPlanYamlArgs.SerializedParams,
			MainFunctionName: req.Msg.StarlarkScriptPlanYamlArgs.MainFunctionName,
		},
	}
	result, err := (*apiContainerServiceClient).GetStarlarkScriptPlanYaml(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.PlanYaml]{
		Msg: &kurtosis_core_rpc_api_bindings.PlanYaml{
			PlanYaml: result.Msg.PlanYaml,
		},
	}
	return resp, nil
}

func (c *WebServer) GetStarlarkPackagePlanYaml(
	ctx context.Context,
	req *connect.Request[kurtosis_enclave_manager_api_bindings.StarlarkPackagePlanYamlArgs],
) (*connect.Response[kurtosis_core_rpc_api_bindings.PlanYaml], error) {
	auth, _, err := c.ValidateRequestAuthorization(ctx, c.enforceAuth, req.Header())
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

	request := &connect.Request[kurtosis_core_rpc_api_bindings.StarlarkPackagePlanYamlArgs]{
		Msg: &kurtosis_core_rpc_api_bindings.StarlarkPackagePlanYamlArgs{
			PackageId:              req.Msg.StarlarkPackagePlanYamlArgs.PackageId,
			SerializedParams:       req.Msg.StarlarkPackagePlanYamlArgs.SerializedParams,
			RelativePathToMainFile: req.Msg.StarlarkPackagePlanYamlArgs.RelativePathToMainFile,
			MainFunctionName:       req.Msg.StarlarkPackagePlanYamlArgs.MainFunctionName,
		},
	}
	result, err := (*apiContainerServiceClient).GetStarlarkPackagePlanYaml(ctx, request)
	if err != nil {
		return nil, err
	}
	resp := &connect.Response[kurtosis_core_rpc_api_bindings.PlanYaml]{
		Msg: &kurtosis_core_rpc_api_bindings.PlanYaml{
			PlanYaml: result.Msg.PlanYaml,
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
	parsedUrl, err := url.Parse(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to parse the connection url for Kurtosis Cloud Backend")
	}
	client := kurtosis_backend_server_rpc_api_bindingsconnect.NewKurtosisCloudBackendServerClient(
		http.DefaultClient,
		parsedUrl.String(),
		connect.WithGRPCWeb(),
	)
	return &client, nil
}

func (c *WebServer) getCloudInstanceConfig(
	ctx context.Context,
	jwtToken string,
	apiKey string,
) (*kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse, error) {
	// Check if we have already fetched the instance config, if so return the cache
	if c.instanceConfigMap[jwtToken] != nil {
		return c.instanceConfigMap[jwtToken], nil
	}

	// We have not yet fetched the instance configuration, so we write lock, make the external call and cache the result
	c.instanceConfigMutex.Lock()
	defer c.instanceConfigMutex.Unlock()

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
	// nolint:exhaustruct
	getInstanceConfigRequest := &connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]{
		Msg: &kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs{
			ApiKey:     &apiKey,
			InstanceId: &getInstanceResponse.Msg.InstanceId,
		},
	}
	getInstanceConfigResponse, err := (*client).GetCloudInstanceConfig(ctx, getInstanceConfigRequest)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get the instance config")
	}

	c.instanceConfigMap[jwtToken] = getInstanceConfigResponse.Msg

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

	if c.apiKeyMap[jwtToken] != nil {
		return &Authentication{
			ApiKey:   *c.apiKeyMap[jwtToken],
			JwtToken: jwtToken,
		}, nil
	} else {
		c.apiKeyMutex.Lock()
		defer c.apiKeyMutex.Unlock()

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
			// User does not have an API key (not really possible if they have a valid jwt token)
			return nil, stacktrace.NewError("User does not have an API key assigned")
		}

		if len(result.Msg.ApiKey) > 0 {
			c.apiKeyMap[jwtToken] = &result.Msg.ApiKey
			return &Authentication{
				ApiKey:   result.Msg.ApiKey,
				JwtToken: jwtToken,
			}, nil
		}
	}

	return nil, stacktrace.NewError("an empty API key was returned from Kurtosis Cloud Backend")
}

func extractJwtToken(header http.Header) (string, error) {
	reqToken := header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != numberOfElementsAuthHeader {
		return "", stacktrace.NewError("Authorization token malformed. Bearer token format required")
	}
	reqToken = strings.TrimSpace(splitToken[1])
	return reqToken, nil
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

	emCors := cors.AllowAll()
	emCors.Log = logrus.StandardLogger()

	if err := apiServer.RunServerUntilInterruptedWithCors(emCors); err != nil {
		logrus.Error("An error occurred running the server", err)
	}
	return nil
}
