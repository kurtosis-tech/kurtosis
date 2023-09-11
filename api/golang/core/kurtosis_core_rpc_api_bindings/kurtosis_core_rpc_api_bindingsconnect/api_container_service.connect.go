// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: api_container_service.proto

package kurtosis_core_rpc_api_bindingsconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion0_1_0

const (
	// ApiContainerServiceName is the fully-qualified name of the ApiContainerService service.
	ApiContainerServiceName = "api_container_api.ApiContainerService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ApiContainerServiceRunStarlarkScriptProcedure is the fully-qualified name of the
	// ApiContainerService's RunStarlarkScript RPC.
	ApiContainerServiceRunStarlarkScriptProcedure = "/api_container_api.ApiContainerService/RunStarlarkScript"
	// ApiContainerServiceUploadStarlarkPackageProcedure is the fully-qualified name of the
	// ApiContainerService's UploadStarlarkPackage RPC.
	ApiContainerServiceUploadStarlarkPackageProcedure = "/api_container_api.ApiContainerService/UploadStarlarkPackage"
	// ApiContainerServiceRunStarlarkPackageProcedure is the fully-qualified name of the
	// ApiContainerService's RunStarlarkPackage RPC.
	ApiContainerServiceRunStarlarkPackageProcedure = "/api_container_api.ApiContainerService/RunStarlarkPackage"
	// ApiContainerServiceGetServicesProcedure is the fully-qualified name of the ApiContainerService's
	// GetServices RPC.
	ApiContainerServiceGetServicesProcedure = "/api_container_api.ApiContainerService/GetServices"
	// ApiContainerServiceGetExistingAndHistoricalServiceIdentifiersProcedure is the fully-qualified
	// name of the ApiContainerService's GetExistingAndHistoricalServiceIdentifiers RPC.
	ApiContainerServiceGetExistingAndHistoricalServiceIdentifiersProcedure = "/api_container_api.ApiContainerService/GetExistingAndHistoricalServiceIdentifiers"
	// ApiContainerServiceExecCommandProcedure is the fully-qualified name of the ApiContainerService's
	// ExecCommand RPC.
	ApiContainerServiceExecCommandProcedure = "/api_container_api.ApiContainerService/ExecCommand"
	// ApiContainerServiceWaitForHttpGetEndpointAvailabilityProcedure is the fully-qualified name of the
	// ApiContainerService's WaitForHttpGetEndpointAvailability RPC.
	ApiContainerServiceWaitForHttpGetEndpointAvailabilityProcedure = "/api_container_api.ApiContainerService/WaitForHttpGetEndpointAvailability"
	// ApiContainerServiceWaitForHttpPostEndpointAvailabilityProcedure is the fully-qualified name of
	// the ApiContainerService's WaitForHttpPostEndpointAvailability RPC.
	ApiContainerServiceWaitForHttpPostEndpointAvailabilityProcedure = "/api_container_api.ApiContainerService/WaitForHttpPostEndpointAvailability"
	// ApiContainerServiceUploadFilesArtifactProcedure is the fully-qualified name of the
	// ApiContainerService's UploadFilesArtifact RPC.
	ApiContainerServiceUploadFilesArtifactProcedure = "/api_container_api.ApiContainerService/UploadFilesArtifact"
	// ApiContainerServiceDownloadFilesArtifactProcedure is the fully-qualified name of the
	// ApiContainerService's DownloadFilesArtifact RPC.
	ApiContainerServiceDownloadFilesArtifactProcedure = "/api_container_api.ApiContainerService/DownloadFilesArtifact"
	// ApiContainerServiceStoreWebFilesArtifactProcedure is the fully-qualified name of the
	// ApiContainerService's StoreWebFilesArtifact RPC.
	ApiContainerServiceStoreWebFilesArtifactProcedure = "/api_container_api.ApiContainerService/StoreWebFilesArtifact"
	// ApiContainerServiceStoreFilesArtifactFromServiceProcedure is the fully-qualified name of the
	// ApiContainerService's StoreFilesArtifactFromService RPC.
	ApiContainerServiceStoreFilesArtifactFromServiceProcedure = "/api_container_api.ApiContainerService/StoreFilesArtifactFromService"
	// ApiContainerServiceListFilesArtifactNamesAndUuidsProcedure is the fully-qualified name of the
	// ApiContainerService's ListFilesArtifactNamesAndUuids RPC.
	ApiContainerServiceListFilesArtifactNamesAndUuidsProcedure = "/api_container_api.ApiContainerService/ListFilesArtifactNamesAndUuids"
	// ApiContainerServiceInspectFilesArtifactContentsProcedure is the fully-qualified name of the
	// ApiContainerService's InspectFilesArtifactContents RPC.
	ApiContainerServiceInspectFilesArtifactContentsProcedure = "/api_container_api.ApiContainerService/InspectFilesArtifactContents"
	// ApiContainerServiceConnectServicesProcedure is the fully-qualified name of the
	// ApiContainerService's ConnectServices RPC.
	ApiContainerServiceConnectServicesProcedure = "/api_container_api.ApiContainerService/ConnectServices"
)

// ApiContainerServiceClient is a client for the api_container_api.ApiContainerService service.
type ApiContainerServiceClient interface {
	// Executes a Starlark script on the user's behalf
	RunStarlarkScript(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error)
	// Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
	UploadStarlarkPackage(context.Context) *connect.ClientStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty]
	// Executes a Starlark script on the user's behalf
	RunStarlarkPackage(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error)
	// Returns the IDs of the current services in the enclave
	GetServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error)
	// Returns information about all existing & historical services
	GetExistingAndHistoricalServiceIdentifiers(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse], error)
	// Executes the given command inside a running container
	ExecCommand(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ExecCommandArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ExecCommandResponse], error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
	WaitForHttpGetEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
	WaitForHttpPostEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error)
	// Uploads a files artifact to the Kurtosis File System
	UploadFilesArtifact(context.Context) *connect.ClientStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse]
	// Downloads a files artifact from the Kurtosis File System
	DownloadFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk], error)
	// Tells the API container to download a files artifact from the web to the Kurtosis File System
	StoreWebFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse], error)
	// Tells the API container to copy a files artifact from a service to the Kurtosis File System
	StoreFilesArtifactFromService(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse], error)
	ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error)
	InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error)
	// User services port forwarding
	ConnectServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ConnectServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ConnectServicesResponse], error)
}

// NewApiContainerServiceClient constructs a client for the api_container_api.ApiContainerService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewApiContainerServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ApiContainerServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &apiContainerServiceClient{
		runStarlarkScript: connect.NewClient[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine](
			httpClient,
			baseURL+ApiContainerServiceRunStarlarkScriptProcedure,
			opts...,
		),
		uploadStarlarkPackage: connect.NewClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty](
			httpClient,
			baseURL+ApiContainerServiceUploadStarlarkPackageProcedure,
			opts...,
		),
		runStarlarkPackage: connect.NewClient[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine](
			httpClient,
			baseURL+ApiContainerServiceRunStarlarkPackageProcedure,
			opts...,
		),
		getServices: connect.NewClient[kurtosis_core_rpc_api_bindings.GetServicesArgs, kurtosis_core_rpc_api_bindings.GetServicesResponse](
			httpClient,
			baseURL+ApiContainerServiceGetServicesProcedure,
			opts...,
		),
		getExistingAndHistoricalServiceIdentifiers: connect.NewClient[emptypb.Empty, kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse](
			httpClient,
			baseURL+ApiContainerServiceGetExistingAndHistoricalServiceIdentifiersProcedure,
			opts...,
		),
		execCommand: connect.NewClient[kurtosis_core_rpc_api_bindings.ExecCommandArgs, kurtosis_core_rpc_api_bindings.ExecCommandResponse](
			httpClient,
			baseURL+ApiContainerServiceExecCommandProcedure,
			opts...,
		),
		waitForHttpGetEndpointAvailability: connect.NewClient[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs, emptypb.Empty](
			httpClient,
			baseURL+ApiContainerServiceWaitForHttpGetEndpointAvailabilityProcedure,
			opts...,
		),
		waitForHttpPostEndpointAvailability: connect.NewClient[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs, emptypb.Empty](
			httpClient,
			baseURL+ApiContainerServiceWaitForHttpPostEndpointAvailabilityProcedure,
			opts...,
		),
		uploadFilesArtifact: connect.NewClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse](
			httpClient,
			baseURL+ApiContainerServiceUploadFilesArtifactProcedure,
			opts...,
		),
		downloadFilesArtifact: connect.NewClient[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs, kurtosis_core_rpc_api_bindings.StreamedDataChunk](
			httpClient,
			baseURL+ApiContainerServiceDownloadFilesArtifactProcedure,
			opts...,
		),
		storeWebFilesArtifact: connect.NewClient[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs, kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse](
			httpClient,
			baseURL+ApiContainerServiceStoreWebFilesArtifactProcedure,
			opts...,
		),
		storeFilesArtifactFromService: connect.NewClient[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs, kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse](
			httpClient,
			baseURL+ApiContainerServiceStoreFilesArtifactFromServiceProcedure,
			opts...,
		),
		listFilesArtifactNamesAndUuids: connect.NewClient[emptypb.Empty, kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse](
			httpClient,
			baseURL+ApiContainerServiceListFilesArtifactNamesAndUuidsProcedure,
			opts...,
		),
		inspectFilesArtifactContents: connect.NewClient[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest, kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse](
			httpClient,
			baseURL+ApiContainerServiceInspectFilesArtifactContentsProcedure,
			opts...,
		),
		connectServices: connect.NewClient[kurtosis_core_rpc_api_bindings.ConnectServicesArgs, kurtosis_core_rpc_api_bindings.ConnectServicesResponse](
			httpClient,
			baseURL+ApiContainerServiceConnectServicesProcedure,
			opts...,
		),
	}
}

// apiContainerServiceClient implements ApiContainerServiceClient.
type apiContainerServiceClient struct {
	runStarlarkScript                          *connect.Client[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]
	uploadStarlarkPackage                      *connect.Client[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty]
	runStarlarkPackage                         *connect.Client[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]
	getServices                                *connect.Client[kurtosis_core_rpc_api_bindings.GetServicesArgs, kurtosis_core_rpc_api_bindings.GetServicesResponse]
	getExistingAndHistoricalServiceIdentifiers *connect.Client[emptypb.Empty, kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse]
	execCommand                                *connect.Client[kurtosis_core_rpc_api_bindings.ExecCommandArgs, kurtosis_core_rpc_api_bindings.ExecCommandResponse]
	waitForHttpGetEndpointAvailability         *connect.Client[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs, emptypb.Empty]
	waitForHttpPostEndpointAvailability        *connect.Client[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs, emptypb.Empty]
	uploadFilesArtifact                        *connect.Client[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse]
	downloadFilesArtifact                      *connect.Client[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs, kurtosis_core_rpc_api_bindings.StreamedDataChunk]
	storeWebFilesArtifact                      *connect.Client[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs, kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse]
	storeFilesArtifactFromService              *connect.Client[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs, kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse]
	listFilesArtifactNamesAndUuids             *connect.Client[emptypb.Empty, kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse]
	inspectFilesArtifactContents               *connect.Client[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest, kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse]
	connectServices                            *connect.Client[kurtosis_core_rpc_api_bindings.ConnectServicesArgs, kurtosis_core_rpc_api_bindings.ConnectServicesResponse]
}

// RunStarlarkScript calls api_container_api.ApiContainerService.RunStarlarkScript.
func (c *apiContainerServiceClient) RunStarlarkScript(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error) {
	return c.runStarlarkScript.CallServerStream(ctx, req)
}

// UploadStarlarkPackage calls api_container_api.ApiContainerService.UploadStarlarkPackage.
func (c *apiContainerServiceClient) UploadStarlarkPackage(ctx context.Context) *connect.ClientStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty] {
	return c.uploadStarlarkPackage.CallClientStream(ctx)
}

// RunStarlarkPackage calls api_container_api.ApiContainerService.RunStarlarkPackage.
func (c *apiContainerServiceClient) RunStarlarkPackage(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error) {
	return c.runStarlarkPackage.CallServerStream(ctx, req)
}

// GetServices calls api_container_api.ApiContainerService.GetServices.
func (c *apiContainerServiceClient) GetServices(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {
	return c.getServices.CallUnary(ctx, req)
}

// GetExistingAndHistoricalServiceIdentifiers calls
// api_container_api.ApiContainerService.GetExistingAndHistoricalServiceIdentifiers.
func (c *apiContainerServiceClient) GetExistingAndHistoricalServiceIdentifiers(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse], error) {
	return c.getExistingAndHistoricalServiceIdentifiers.CallUnary(ctx, req)
}

// ExecCommand calls api_container_api.ApiContainerService.ExecCommand.
func (c *apiContainerServiceClient) ExecCommand(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.ExecCommandArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ExecCommandResponse], error) {
	return c.execCommand.CallUnary(ctx, req)
}

// WaitForHttpGetEndpointAvailability calls
// api_container_api.ApiContainerService.WaitForHttpGetEndpointAvailability.
func (c *apiContainerServiceClient) WaitForHttpGetEndpointAvailability(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.waitForHttpGetEndpointAvailability.CallUnary(ctx, req)
}

// WaitForHttpPostEndpointAvailability calls
// api_container_api.ApiContainerService.WaitForHttpPostEndpointAvailability.
func (c *apiContainerServiceClient) WaitForHttpPostEndpointAvailability(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.waitForHttpPostEndpointAvailability.CallUnary(ctx, req)
}

// UploadFilesArtifact calls api_container_api.ApiContainerService.UploadFilesArtifact.
func (c *apiContainerServiceClient) UploadFilesArtifact(ctx context.Context) *connect.ClientStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse] {
	return c.uploadFilesArtifact.CallClientStream(ctx)
}

// DownloadFilesArtifact calls api_container_api.ApiContainerService.DownloadFilesArtifact.
func (c *apiContainerServiceClient) DownloadFilesArtifact(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StreamedDataChunk], error) {
	return c.downloadFilesArtifact.CallServerStream(ctx, req)
}

// StoreWebFilesArtifact calls api_container_api.ApiContainerService.StoreWebFilesArtifact.
func (c *apiContainerServiceClient) StoreWebFilesArtifact(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse], error) {
	return c.storeWebFilesArtifact.CallUnary(ctx, req)
}

// StoreFilesArtifactFromService calls
// api_container_api.ApiContainerService.StoreFilesArtifactFromService.
func (c *apiContainerServiceClient) StoreFilesArtifactFromService(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse], error) {
	return c.storeFilesArtifactFromService.CallUnary(ctx, req)
}

// ListFilesArtifactNamesAndUuids calls
// api_container_api.ApiContainerService.ListFilesArtifactNamesAndUuids.
func (c *apiContainerServiceClient) ListFilesArtifactNamesAndUuids(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	return c.listFilesArtifactNamesAndUuids.CallUnary(ctx, req)
}

// InspectFilesArtifactContents calls
// api_container_api.ApiContainerService.InspectFilesArtifactContents.
func (c *apiContainerServiceClient) InspectFilesArtifactContents(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	return c.inspectFilesArtifactContents.CallUnary(ctx, req)
}

// ConnectServices calls api_container_api.ApiContainerService.ConnectServices.
func (c *apiContainerServiceClient) ConnectServices(ctx context.Context, req *connect.Request[kurtosis_core_rpc_api_bindings.ConnectServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ConnectServicesResponse], error) {
	return c.connectServices.CallUnary(ctx, req)
}

// ApiContainerServiceHandler is an implementation of the api_container_api.ApiContainerService
// service.
type ApiContainerServiceHandler interface {
	// Executes a Starlark script on the user's behalf
	RunStarlarkScript(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error
	// Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
	UploadStarlarkPackage(context.Context, *connect.ClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) (*connect.Response[emptypb.Empty], error)
	// Executes a Starlark script on the user's behalf
	RunStarlarkPackage(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error
	// Returns the IDs of the current services in the enclave
	GetServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error)
	// Returns information about all existing & historical services
	GetExistingAndHistoricalServiceIdentifiers(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse], error)
	// Executes the given command inside a running container
	ExecCommand(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ExecCommandArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ExecCommandResponse], error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
	WaitForHttpGetEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error)
	// Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
	WaitForHttpPostEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error)
	// Uploads a files artifact to the Kurtosis File System
	UploadFilesArtifact(context.Context, *connect.ClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) (*connect.Response[kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse], error)
	// Downloads a files artifact from the Kurtosis File System
	DownloadFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) error
	// Tells the API container to download a files artifact from the web to the Kurtosis File System
	StoreWebFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse], error)
	// Tells the API container to copy a files artifact from a service to the Kurtosis File System
	StoreFilesArtifactFromService(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse], error)
	ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error)
	InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error)
	// User services port forwarding
	ConnectServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ConnectServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ConnectServicesResponse], error)
}

// NewApiContainerServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewApiContainerServiceHandler(svc ApiContainerServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	apiContainerServiceRunStarlarkScriptHandler := connect.NewServerStreamHandler(
		ApiContainerServiceRunStarlarkScriptProcedure,
		svc.RunStarlarkScript,
		opts...,
	)
	apiContainerServiceUploadStarlarkPackageHandler := connect.NewClientStreamHandler(
		ApiContainerServiceUploadStarlarkPackageProcedure,
		svc.UploadStarlarkPackage,
		opts...,
	)
	apiContainerServiceRunStarlarkPackageHandler := connect.NewServerStreamHandler(
		ApiContainerServiceRunStarlarkPackageProcedure,
		svc.RunStarlarkPackage,
		opts...,
	)
	apiContainerServiceGetServicesHandler := connect.NewUnaryHandler(
		ApiContainerServiceGetServicesProcedure,
		svc.GetServices,
		opts...,
	)
	apiContainerServiceGetExistingAndHistoricalServiceIdentifiersHandler := connect.NewUnaryHandler(
		ApiContainerServiceGetExistingAndHistoricalServiceIdentifiersProcedure,
		svc.GetExistingAndHistoricalServiceIdentifiers,
		opts...,
	)
	apiContainerServiceExecCommandHandler := connect.NewUnaryHandler(
		ApiContainerServiceExecCommandProcedure,
		svc.ExecCommand,
		opts...,
	)
	apiContainerServiceWaitForHttpGetEndpointAvailabilityHandler := connect.NewUnaryHandler(
		ApiContainerServiceWaitForHttpGetEndpointAvailabilityProcedure,
		svc.WaitForHttpGetEndpointAvailability,
		opts...,
	)
	apiContainerServiceWaitForHttpPostEndpointAvailabilityHandler := connect.NewUnaryHandler(
		ApiContainerServiceWaitForHttpPostEndpointAvailabilityProcedure,
		svc.WaitForHttpPostEndpointAvailability,
		opts...,
	)
	apiContainerServiceUploadFilesArtifactHandler := connect.NewClientStreamHandler(
		ApiContainerServiceUploadFilesArtifactProcedure,
		svc.UploadFilesArtifact,
		opts...,
	)
	apiContainerServiceDownloadFilesArtifactHandler := connect.NewServerStreamHandler(
		ApiContainerServiceDownloadFilesArtifactProcedure,
		svc.DownloadFilesArtifact,
		opts...,
	)
	apiContainerServiceStoreWebFilesArtifactHandler := connect.NewUnaryHandler(
		ApiContainerServiceStoreWebFilesArtifactProcedure,
		svc.StoreWebFilesArtifact,
		opts...,
	)
	apiContainerServiceStoreFilesArtifactFromServiceHandler := connect.NewUnaryHandler(
		ApiContainerServiceStoreFilesArtifactFromServiceProcedure,
		svc.StoreFilesArtifactFromService,
		opts...,
	)
	apiContainerServiceListFilesArtifactNamesAndUuidsHandler := connect.NewUnaryHandler(
		ApiContainerServiceListFilesArtifactNamesAndUuidsProcedure,
		svc.ListFilesArtifactNamesAndUuids,
		opts...,
	)
	apiContainerServiceInspectFilesArtifactContentsHandler := connect.NewUnaryHandler(
		ApiContainerServiceInspectFilesArtifactContentsProcedure,
		svc.InspectFilesArtifactContents,
		opts...,
	)
	apiContainerServiceConnectServicesHandler := connect.NewUnaryHandler(
		ApiContainerServiceConnectServicesProcedure,
		svc.ConnectServices,
		opts...,
	)
	return "/api_container_api.ApiContainerService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ApiContainerServiceRunStarlarkScriptProcedure:
			apiContainerServiceRunStarlarkScriptHandler.ServeHTTP(w, r)
		case ApiContainerServiceUploadStarlarkPackageProcedure:
			apiContainerServiceUploadStarlarkPackageHandler.ServeHTTP(w, r)
		case ApiContainerServiceRunStarlarkPackageProcedure:
			apiContainerServiceRunStarlarkPackageHandler.ServeHTTP(w, r)
		case ApiContainerServiceGetServicesProcedure:
			apiContainerServiceGetServicesHandler.ServeHTTP(w, r)
		case ApiContainerServiceGetExistingAndHistoricalServiceIdentifiersProcedure:
			apiContainerServiceGetExistingAndHistoricalServiceIdentifiersHandler.ServeHTTP(w, r)
		case ApiContainerServiceExecCommandProcedure:
			apiContainerServiceExecCommandHandler.ServeHTTP(w, r)
		case ApiContainerServiceWaitForHttpGetEndpointAvailabilityProcedure:
			apiContainerServiceWaitForHttpGetEndpointAvailabilityHandler.ServeHTTP(w, r)
		case ApiContainerServiceWaitForHttpPostEndpointAvailabilityProcedure:
			apiContainerServiceWaitForHttpPostEndpointAvailabilityHandler.ServeHTTP(w, r)
		case ApiContainerServiceUploadFilesArtifactProcedure:
			apiContainerServiceUploadFilesArtifactHandler.ServeHTTP(w, r)
		case ApiContainerServiceDownloadFilesArtifactProcedure:
			apiContainerServiceDownloadFilesArtifactHandler.ServeHTTP(w, r)
		case ApiContainerServiceStoreWebFilesArtifactProcedure:
			apiContainerServiceStoreWebFilesArtifactHandler.ServeHTTP(w, r)
		case ApiContainerServiceStoreFilesArtifactFromServiceProcedure:
			apiContainerServiceStoreFilesArtifactFromServiceHandler.ServeHTTP(w, r)
		case ApiContainerServiceListFilesArtifactNamesAndUuidsProcedure:
			apiContainerServiceListFilesArtifactNamesAndUuidsHandler.ServeHTTP(w, r)
		case ApiContainerServiceInspectFilesArtifactContentsProcedure:
			apiContainerServiceInspectFilesArtifactContentsHandler.ServeHTTP(w, r)
		case ApiContainerServiceConnectServicesProcedure:
			apiContainerServiceConnectServicesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedApiContainerServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedApiContainerServiceHandler struct{}

func (UnimplementedApiContainerServiceHandler) RunStarlarkScript(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.RunStarlarkScript is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) UploadStarlarkPackage(context.Context, *connect.ClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.UploadStarlarkPackage is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) RunStarlarkPackage(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.RunStarlarkPackage is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) GetServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.GetServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.GetServices is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) GetExistingAndHistoricalServiceIdentifiers(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.GetExistingAndHistoricalServiceIdentifiers is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) ExecCommand(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ExecCommandArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ExecCommandResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.ExecCommand is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) WaitForHttpGetEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.WaitForHttpGetEndpointAvailability is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) WaitForHttpPostEndpointAvailability(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.WaitForHttpPostEndpointAvailability is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) UploadFilesArtifact(context.Context, *connect.ClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) (*connect.Response[kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.UploadFilesArtifact is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) DownloadFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.DownloadFilesArtifact is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) StoreWebFilesArtifact(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.StoreWebFilesArtifact is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) StoreFilesArtifactFromService(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.StoreFilesArtifactFromService is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.ListFilesArtifactNamesAndUuids is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.InspectFilesArtifactContents is not implemented"))
}

func (UnimplementedApiContainerServiceHandler) ConnectServices(context.Context, *connect.Request[kurtosis_core_rpc_api_bindings.ConnectServicesArgs]) (*connect.Response[kurtosis_core_rpc_api_bindings.ConnectServicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("api_container_api.ApiContainerService.ConnectServices is not implemented"))
}
