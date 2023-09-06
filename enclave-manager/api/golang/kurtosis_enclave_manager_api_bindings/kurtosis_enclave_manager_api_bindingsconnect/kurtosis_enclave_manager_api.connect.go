// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: kurtosis_enclave_manager_api.proto

package kurtosis_enclave_manager_api_bindingsconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	kurtosis_engine_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	kurtosis_enclave_manager_api_bindings "github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings"
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
	// KurtosisEnclaveManagerServerName is the fully-qualified name of the KurtosisEnclaveManagerServer
	// service.
	KurtosisEnclaveManagerServerName = "kurtosis_enclave_manager.KurtosisEnclaveManagerServer"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// KurtosisEnclaveManagerServerCheckProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's Check RPC.
	KurtosisEnclaveManagerServerCheckProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/Check"
	// KurtosisEnclaveManagerServerGetEnclavesProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's GetEnclaves RPC.
	KurtosisEnclaveManagerServerGetEnclavesProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetEnclaves"
	// KurtosisEnclaveManagerServerGetServicesProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's GetServices RPC.
	KurtosisEnclaveManagerServerGetServicesProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetServices"
	// KurtosisEnclaveManagerServerGetServiceLogsProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's GetServiceLogs RPC.
	KurtosisEnclaveManagerServerGetServiceLogsProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetServiceLogs"
	// KurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsProcedure is the fully-qualified name
	// of the KurtosisEnclaveManagerServer's ListFilesArtifactNamesAndUuids RPC.
	KurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/ListFilesArtifactNamesAndUuids"
	// KurtosisEnclaveManagerServerRunStarlarkPackageProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's RunStarlarkPackage RPC.
	KurtosisEnclaveManagerServerRunStarlarkPackageProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/RunStarlarkPackage"
	// KurtosisEnclaveManagerServerCreateEnclaveProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's CreateEnclave RPC.
	KurtosisEnclaveManagerServerCreateEnclaveProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/CreateEnclave"
	// KurtosisEnclaveManagerServerInspectFilesArtifactContentsProcedure is the fully-qualified name of
	// the KurtosisEnclaveManagerServer's InspectFilesArtifactContents RPC.
	KurtosisEnclaveManagerServerInspectFilesArtifactContentsProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/InspectFilesArtifactContents"
	// KurtosisEnclaveManagerServerDestroyEnclaveProcedure is the fully-qualified name of the
	// KurtosisEnclaveManagerServer's DestroyEnclave RPC.
	KurtosisEnclaveManagerServerDestroyEnclaveProcedure = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/DestroyEnclave"
)

// KurtosisEnclaveManagerServerClient is a client for the
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer service.
type KurtosisEnclaveManagerServerClient interface {
	Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error)
	GetEnclaves(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error)
	GetServices(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error)
	GetServiceLogs(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs]) (*connect.ServerStreamForClient[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse], error)
	ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error)
	RunStarlarkPackage(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error)
	CreateEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error)
	InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error)
	DestroyEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error)
}

// NewKurtosisEnclaveManagerServerClient constructs a client for the
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer service. By default, it uses the Connect
// protocol with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed
// requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewKurtosisEnclaveManagerServerClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) KurtosisEnclaveManagerServerClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &kurtosisEnclaveManagerServerClient{
		check: connect.NewClient[kurtosis_enclave_manager_api_bindings.HealthCheckRequest, kurtosis_enclave_manager_api_bindings.HealthCheckResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerCheckProcedure,
			opts...,
		),
		getEnclaves: connect.NewClient[emptypb.Empty, kurtosis_engine_rpc_api_bindings.GetEnclavesResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerGetEnclavesProcedure,
			opts...,
		),
		getServices: connect.NewClient[kurtosis_enclave_manager_api_bindings.GetServicesRequest, kurtosis_core_rpc_api_bindings.GetServicesResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerGetServicesProcedure,
			opts...,
		),
		getServiceLogs: connect.NewClient[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs, kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerGetServiceLogsProcedure,
			opts...,
		),
		listFilesArtifactNamesAndUuids: connect.NewClient[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest, kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsProcedure,
			opts...,
		),
		runStarlarkPackage: connect.NewClient[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerRunStarlarkPackageProcedure,
			opts...,
		),
		createEnclave: connect.NewClient[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs, kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerCreateEnclaveProcedure,
			opts...,
		),
		inspectFilesArtifactContents: connect.NewClient[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest, kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerInspectFilesArtifactContentsProcedure,
			opts...,
		),
		destroyEnclave: connect.NewClient[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs, emptypb.Empty](
			httpClient,
			baseURL+KurtosisEnclaveManagerServerDestroyEnclaveProcedure,
			opts...,
		),
	}
}

// kurtosisEnclaveManagerServerClient implements KurtosisEnclaveManagerServerClient.
type kurtosisEnclaveManagerServerClient struct {
	check                          *connect.Client[kurtosis_enclave_manager_api_bindings.HealthCheckRequest, kurtosis_enclave_manager_api_bindings.HealthCheckResponse]
	getEnclaves                    *connect.Client[emptypb.Empty, kurtosis_engine_rpc_api_bindings.GetEnclavesResponse]
	getServices                    *connect.Client[kurtosis_enclave_manager_api_bindings.GetServicesRequest, kurtosis_core_rpc_api_bindings.GetServicesResponse]
	getServiceLogs                 *connect.Client[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs, kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]
	listFilesArtifactNamesAndUuids *connect.Client[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest, kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse]
	runStarlarkPackage             *connect.Client[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest, kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]
	createEnclave                  *connect.Client[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs, kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse]
	inspectFilesArtifactContents   *connect.Client[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest, kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse]
	destroyEnclave                 *connect.Client[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs, emptypb.Empty]
}

// Check calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.Check.
func (c *kurtosisEnclaveManagerServerClient) Check(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	return c.check.CallUnary(ctx, req)
}

// GetEnclaves calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetEnclaves.
func (c *kurtosisEnclaveManagerServerClient) GetEnclaves(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	return c.getEnclaves.CallUnary(ctx, req)
}

// GetServices calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServices.
func (c *kurtosisEnclaveManagerServerClient) GetServices(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {
	return c.getServices.CallUnary(ctx, req)
}

// GetServiceLogs calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServiceLogs.
func (c *kurtosisEnclaveManagerServerClient) GetServiceLogs(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs]) (*connect.ServerStreamForClient[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse], error) {
	return c.getServiceLogs.CallServerStream(ctx, req)
}

// ListFilesArtifactNamesAndUuids calls
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer.ListFilesArtifactNamesAndUuids.
func (c *kurtosisEnclaveManagerServerClient) ListFilesArtifactNamesAndUuids(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	return c.listFilesArtifactNamesAndUuids.CallUnary(ctx, req)
}

// RunStarlarkPackage calls
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer.RunStarlarkPackage.
func (c *kurtosisEnclaveManagerServerClient) RunStarlarkPackage(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest]) (*connect.ServerStreamForClient[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine], error) {
	return c.runStarlarkPackage.CallServerStream(ctx, req)
}

// CreateEnclave calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.CreateEnclave.
func (c *kurtosisEnclaveManagerServerClient) CreateEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	return c.createEnclave.CallUnary(ctx, req)
}

// InspectFilesArtifactContents calls
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer.InspectFilesArtifactContents.
func (c *kurtosisEnclaveManagerServerClient) InspectFilesArtifactContents(ctx context.Context, req *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	return c.inspectFilesArtifactContents.CallUnary(ctx, req)
}

// DestroyEnclave calls kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DestroyEnclave.
func (c *kurtosisEnclaveManagerServerClient) DestroyEnclave(ctx context.Context, req *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.destroyEnclave.CallUnary(ctx, req)
}

// KurtosisEnclaveManagerServerHandler is an implementation of the
// kurtosis_enclave_manager.KurtosisEnclaveManagerServer service.
type KurtosisEnclaveManagerServerHandler interface {
	Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error)
	GetEnclaves(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error)
	GetServices(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error)
	GetServiceLogs(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs], *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]) error
	ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error)
	RunStarlarkPackage(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error
	CreateEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error)
	InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error)
	DestroyEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error)
}

// NewKurtosisEnclaveManagerServerHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewKurtosisEnclaveManagerServerHandler(svc KurtosisEnclaveManagerServerHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	kurtosisEnclaveManagerServerCheckHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerCheckProcedure,
		svc.Check,
		opts...,
	)
	kurtosisEnclaveManagerServerGetEnclavesHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerGetEnclavesProcedure,
		svc.GetEnclaves,
		opts...,
	)
	kurtosisEnclaveManagerServerGetServicesHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerGetServicesProcedure,
		svc.GetServices,
		opts...,
	)
	kurtosisEnclaveManagerServerGetServiceLogsHandler := connect.NewServerStreamHandler(
		KurtosisEnclaveManagerServerGetServiceLogsProcedure,
		svc.GetServiceLogs,
		opts...,
	)
	kurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsProcedure,
		svc.ListFilesArtifactNamesAndUuids,
		opts...,
	)
	kurtosisEnclaveManagerServerRunStarlarkPackageHandler := connect.NewServerStreamHandler(
		KurtosisEnclaveManagerServerRunStarlarkPackageProcedure,
		svc.RunStarlarkPackage,
		opts...,
	)
	kurtosisEnclaveManagerServerCreateEnclaveHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerCreateEnclaveProcedure,
		svc.CreateEnclave,
		opts...,
	)
	kurtosisEnclaveManagerServerInspectFilesArtifactContentsHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerInspectFilesArtifactContentsProcedure,
		svc.InspectFilesArtifactContents,
		opts...,
	)
	kurtosisEnclaveManagerServerDestroyEnclaveHandler := connect.NewUnaryHandler(
		KurtosisEnclaveManagerServerDestroyEnclaveProcedure,
		svc.DestroyEnclave,
		opts...,
	)
	return "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case KurtosisEnclaveManagerServerCheckProcedure:
			kurtosisEnclaveManagerServerCheckHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerGetEnclavesProcedure:
			kurtosisEnclaveManagerServerGetEnclavesHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerGetServicesProcedure:
			kurtosisEnclaveManagerServerGetServicesHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerGetServiceLogsProcedure:
			kurtosisEnclaveManagerServerGetServiceLogsHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsProcedure:
			kurtosisEnclaveManagerServerListFilesArtifactNamesAndUuidsHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerRunStarlarkPackageProcedure:
			kurtosisEnclaveManagerServerRunStarlarkPackageHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerCreateEnclaveProcedure:
			kurtosisEnclaveManagerServerCreateEnclaveHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerInspectFilesArtifactContentsProcedure:
			kurtosisEnclaveManagerServerInspectFilesArtifactContentsHandler.ServeHTTP(w, r)
		case KurtosisEnclaveManagerServerDestroyEnclaveProcedure:
			kurtosisEnclaveManagerServerDestroyEnclaveHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedKurtosisEnclaveManagerServerHandler returns CodeUnimplemented from all methods.
type UnimplementedKurtosisEnclaveManagerServerHandler struct{}

func (UnimplementedKurtosisEnclaveManagerServerHandler) Check(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.HealthCheckRequest]) (*connect.Response[kurtosis_enclave_manager_api_bindings.HealthCheckResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.Check is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) GetEnclaves(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetEnclaves is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) GetServices(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetServicesRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.GetServicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServices is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) GetServiceLogs(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs], *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServiceLogs is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) ListFilesArtifactNamesAndUuids(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.GetListFilesArtifactNamesAndUuidsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.ListFilesArtifactNamesAndUuids is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) RunStarlarkPackage(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.RunStarlarkPackageRequest], *connect.ServerStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.RunStarlarkPackage is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) CreateEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.CreateEnclave is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) InspectFilesArtifactContents(context.Context, *connect.Request[kurtosis_enclave_manager_api_bindings.InspectFilesArtifactContentsRequest]) (*connect.Response[kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.InspectFilesArtifactContents is not implemented"))
}

func (UnimplementedKurtosisEnclaveManagerServerHandler) DestroyEnclave(context.Context, *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DestroyEnclave is not implemented"))
}
