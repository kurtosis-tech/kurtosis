// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: kurtosis_backend_server_api.proto

package kurtosis_backend_server_rpc_api_bindingsconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	kurtosis_backend_server_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings"
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
	// KurtosisCloudBackendServerName is the fully-qualified name of the KurtosisCloudBackendServer
	// service.
	KurtosisCloudBackendServerName = "kurtosis_cloud.KurtosisCloudBackendServer"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// KurtosisCloudBackendServerIsAvailableProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's IsAvailable RPC.
	KurtosisCloudBackendServerIsAvailableProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/IsAvailable"
	// KurtosisCloudBackendServerCreateCloudInstanceProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's CreateCloudInstance RPC.
	KurtosisCloudBackendServerCreateCloudInstanceProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/CreateCloudInstance"
	// KurtosisCloudBackendServerGetCloudInstanceConfigProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetCloudInstanceConfig RPC.
	KurtosisCloudBackendServerGetCloudInstanceConfigProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetCloudInstanceConfig"
	// KurtosisCloudBackendServerGetOrCreateApiKeyProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetOrCreateApiKey RPC.
	KurtosisCloudBackendServerGetOrCreateApiKeyProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetOrCreateApiKey"
	// KurtosisCloudBackendServerGetOrCreateInstanceProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetOrCreateInstance RPC.
	KurtosisCloudBackendServerGetOrCreateInstanceProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetOrCreateInstance"
	// KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetOrCreatePaymentConfig RPC.
	KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetOrCreatePaymentConfig"
)

// KurtosisCloudBackendServerClient is a client for the kurtosis_cloud.KurtosisCloudBackendServer
// service.
type KurtosisCloudBackendServerClient interface {
	IsAvailable(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error)
	CreateCloudInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse], error)
	GetCloudInstanceConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error)
	GetOrCreateApiKey(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error)
	GetOrCreateInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error)
	GetOrCreatePaymentConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error)
}

// NewKurtosisCloudBackendServerClient constructs a client for the
// kurtosis_cloud.KurtosisCloudBackendServer service. By default, it uses the Connect protocol with
// the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To use
// the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewKurtosisCloudBackendServerClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) KurtosisCloudBackendServerClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &kurtosisCloudBackendServerClient{
		isAvailable: connect.NewClient[emptypb.Empty, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerIsAvailableProcedure,
			opts...,
		),
		createCloudInstance: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerCreateCloudInstanceProcedure,
			opts...,
		),
		getCloudInstanceConfig: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetCloudInstanceConfigProcedure,
			opts...,
		),
		getOrCreateApiKey: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetOrCreateApiKeyProcedure,
			opts...,
		),
		getOrCreateInstance: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetOrCreateInstanceProcedure,
			opts...,
		),
		getOrCreatePaymentConfig: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure,
			opts...,
		),
	}
}

// kurtosisCloudBackendServerClient implements KurtosisCloudBackendServerClient.
type kurtosisCloudBackendServerClient struct {
	isAvailable              *connect.Client[emptypb.Empty, emptypb.Empty]
	createCloudInstance      *connect.Client[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse]
	getCloudInstanceConfig   *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse]
	getOrCreateApiKey        *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse]
	getOrCreateInstance      *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse]
	getOrCreatePaymentConfig *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse]
}

// IsAvailable calls kurtosis_cloud.KurtosisCloudBackendServer.IsAvailable.
func (c *kurtosisCloudBackendServerClient) IsAvailable(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return c.isAvailable.CallUnary(ctx, req)
}

// CreateCloudInstance calls kurtosis_cloud.KurtosisCloudBackendServer.CreateCloudInstance.
func (c *kurtosisCloudBackendServerClient) CreateCloudInstance(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse], error) {
	return c.createCloudInstance.CallUnary(ctx, req)
}

// GetCloudInstanceConfig calls kurtosis_cloud.KurtosisCloudBackendServer.GetCloudInstanceConfig.
func (c *kurtosisCloudBackendServerClient) GetCloudInstanceConfig(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error) {
	return c.getCloudInstanceConfig.CallUnary(ctx, req)
}

// GetOrCreateApiKey calls kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreateApiKey.
func (c *kurtosisCloudBackendServerClient) GetOrCreateApiKey(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error) {
	return c.getOrCreateApiKey.CallUnary(ctx, req)
}

// GetOrCreateInstance calls kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreateInstance.
func (c *kurtosisCloudBackendServerClient) GetOrCreateInstance(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error) {
	return c.getOrCreateInstance.CallUnary(ctx, req)
}

// GetOrCreatePaymentConfig calls
// kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreatePaymentConfig.
func (c *kurtosisCloudBackendServerClient) GetOrCreatePaymentConfig(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error) {
	return c.getOrCreatePaymentConfig.CallUnary(ctx, req)
}

// KurtosisCloudBackendServerHandler is an implementation of the
// kurtosis_cloud.KurtosisCloudBackendServer service.
type KurtosisCloudBackendServerHandler interface {
	IsAvailable(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error)
	CreateCloudInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse], error)
	GetCloudInstanceConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error)
	GetOrCreateApiKey(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error)
	GetOrCreateInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error)
	GetOrCreatePaymentConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error)
}

// NewKurtosisCloudBackendServerHandler builds an HTTP handler from the service implementation. It
// returns the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewKurtosisCloudBackendServerHandler(svc KurtosisCloudBackendServerHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	kurtosisCloudBackendServerIsAvailableHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerIsAvailableProcedure,
		svc.IsAvailable,
		opts...,
	)
	kurtosisCloudBackendServerCreateCloudInstanceHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerCreateCloudInstanceProcedure,
		svc.CreateCloudInstance,
		opts...,
	)
	kurtosisCloudBackendServerGetCloudInstanceConfigHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetCloudInstanceConfigProcedure,
		svc.GetCloudInstanceConfig,
		opts...,
	)
	kurtosisCloudBackendServerGetOrCreateApiKeyHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetOrCreateApiKeyProcedure,
		svc.GetOrCreateApiKey,
		opts...,
	)
	kurtosisCloudBackendServerGetOrCreateInstanceHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetOrCreateInstanceProcedure,
		svc.GetOrCreateInstance,
		opts...,
	)
	kurtosisCloudBackendServerGetOrCreatePaymentConfigHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure,
		svc.GetOrCreatePaymentConfig,
		opts...,
	)
	return "/kurtosis_cloud.KurtosisCloudBackendServer/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case KurtosisCloudBackendServerIsAvailableProcedure:
			kurtosisCloudBackendServerIsAvailableHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerCreateCloudInstanceProcedure:
			kurtosisCloudBackendServerCreateCloudInstanceHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetCloudInstanceConfigProcedure:
			kurtosisCloudBackendServerGetCloudInstanceConfigHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreateApiKeyProcedure:
			kurtosisCloudBackendServerGetOrCreateApiKeyHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreateInstanceProcedure:
			kurtosisCloudBackendServerGetOrCreateInstanceHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure:
			kurtosisCloudBackendServerGetOrCreatePaymentConfigHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedKurtosisCloudBackendServerHandler returns CodeUnimplemented from all methods.
type UnimplementedKurtosisCloudBackendServerHandler struct{}

func (UnimplementedKurtosisCloudBackendServerHandler) IsAvailable(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.IsAvailable is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) CreateCloudInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.CreateCloudInstanceConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.CreateCloudInstance is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetCloudInstanceConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetCloudInstanceConfig is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetOrCreateApiKey(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreateApiKey is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetOrCreateInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreateInstance is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetOrCreatePaymentConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetOrCreatePaymentConfig is not implemented"))
}
