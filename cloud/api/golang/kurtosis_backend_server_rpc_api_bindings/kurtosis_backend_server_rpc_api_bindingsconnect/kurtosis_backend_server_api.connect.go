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
const _ = connect.IsAtLeastVersion1_7_0

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
	// KurtosisCloudBackendServerRefreshDefaultPaymentMethodProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's RefreshDefaultPaymentMethod RPC.
	KurtosisCloudBackendServerRefreshDefaultPaymentMethodProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/RefreshDefaultPaymentMethod"
	// KurtosisCloudBackendServerCancelPaymentSubscriptionProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's CancelPaymentSubscription RPC.
	KurtosisCloudBackendServerCancelPaymentSubscriptionProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/CancelPaymentSubscription"
	// KurtosisCloudBackendServerUpdateAddressProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's UpdateAddress RPC.
	KurtosisCloudBackendServerUpdateAddressProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/UpdateAddress"
	// KurtosisCloudBackendServerGetInstancesProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetInstances RPC.
	KurtosisCloudBackendServerGetInstancesProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetInstances"
	// KurtosisCloudBackendServerDeleteInstanceProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's DeleteInstance RPC.
	KurtosisCloudBackendServerDeleteInstanceProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/DeleteInstance"
	// KurtosisCloudBackendServerChangeActiveStatusProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's ChangeActiveStatus RPC.
	KurtosisCloudBackendServerChangeActiveStatusProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/ChangeActiveStatus"
	// KurtosisCloudBackendServerGetUserProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetUser RPC.
	KurtosisCloudBackendServerGetUserProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetUser"
	// KurtosisCloudBackendServerCheckPortAuthorizationProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's CheckPortAuthorization RPC.
	KurtosisCloudBackendServerCheckPortAuthorizationProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/CheckPortAuthorization"
	// KurtosisCloudBackendServerUnlockPortProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's UnlockPort RPC.
	KurtosisCloudBackendServerUnlockPortProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/UnlockPort"
	// KurtosisCloudBackendServerLockPortProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's LockPort RPC.
	KurtosisCloudBackendServerLockPortProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/LockPort"
	// KurtosisCloudBackendServerGetUnlockedPortsProcedure is the fully-qualified name of the
	// KurtosisCloudBackendServer's GetUnlockedPorts RPC.
	KurtosisCloudBackendServerGetUnlockedPortsProcedure = "/kurtosis_cloud.KurtosisCloudBackendServer/GetUnlockedPorts"
)

// KurtosisCloudBackendServerClient is a client for the kurtosis_cloud.KurtosisCloudBackendServer
// service.
type KurtosisCloudBackendServerClient interface {
	IsAvailable(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error)
	GetCloudInstanceConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error)
	GetOrCreateApiKey(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error)
	GetOrCreateInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error)
	GetOrCreatePaymentConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error)
	RefreshDefaultPaymentMethod(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs]) (*connect.Response[emptypb.Empty], error)
	CancelPaymentSubscription(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs]) (*connect.Response[emptypb.Empty], error)
	UpdateAddress(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs]) (*connect.Response[emptypb.Empty], error)
	GetInstances(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse], error)
	DeleteInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse], error)
	ChangeActiveStatus(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest]) (*connect.Response[emptypb.Empty], error)
	GetUser(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUserRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUserResponse], error)
	CheckPortAuthorization(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest]) (*connect.Response[emptypb.Empty], error)
	UnlockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest]) (*connect.Response[emptypb.Empty], error)
	LockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.LockPortRequest]) (*connect.Response[emptypb.Empty], error)
	GetUnlockedPorts(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse], error)
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
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		getCloudInstanceConfig: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetCloudInstanceConfigProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
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
		refreshDefaultPaymentMethod: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerRefreshDefaultPaymentMethodProcedure,
			opts...,
		),
		cancelPaymentSubscription: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerCancelPaymentSubscriptionProcedure,
			opts...,
		),
		updateAddress: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerUpdateAddressProcedure,
			opts...,
		),
		getInstances: connect.NewClient[emptypb.Empty, kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetInstancesProcedure,
			opts...,
		),
		deleteInstance: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest, kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerDeleteInstanceProcedure,
			opts...,
		),
		changeActiveStatus: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerChangeActiveStatusProcedure,
			opts...,
		),
		getUser: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetUserRequest, kurtosis_backend_server_rpc_api_bindings.GetUserResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetUserProcedure,
			opts...,
		),
		checkPortAuthorization: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerCheckPortAuthorizationProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		unlockPort: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerUnlockPortProcedure,
			opts...,
		),
		lockPort: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.LockPortRequest, emptypb.Empty](
			httpClient,
			baseURL+KurtosisCloudBackendServerLockPortProcedure,
			opts...,
		),
		getUnlockedPorts: connect.NewClient[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest, kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse](
			httpClient,
			baseURL+KurtosisCloudBackendServerGetUnlockedPortsProcedure,
			opts...,
		),
	}
}

// kurtosisCloudBackendServerClient implements KurtosisCloudBackendServerClient.
type kurtosisCloudBackendServerClient struct {
	isAvailable                 *connect.Client[emptypb.Empty, emptypb.Empty]
	getCloudInstanceConfig      *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse]
	getOrCreateApiKey           *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse]
	getOrCreateInstance         *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest, kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse]
	getOrCreatePaymentConfig    *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs, kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse]
	refreshDefaultPaymentMethod *connect.Client[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs, emptypb.Empty]
	cancelPaymentSubscription   *connect.Client[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs, emptypb.Empty]
	updateAddress               *connect.Client[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs, emptypb.Empty]
	getInstances                *connect.Client[emptypb.Empty, kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse]
	deleteInstance              *connect.Client[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest, kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse]
	changeActiveStatus          *connect.Client[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest, emptypb.Empty]
	getUser                     *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetUserRequest, kurtosis_backend_server_rpc_api_bindings.GetUserResponse]
	checkPortAuthorization      *connect.Client[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest, emptypb.Empty]
	unlockPort                  *connect.Client[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest, emptypb.Empty]
	lockPort                    *connect.Client[kurtosis_backend_server_rpc_api_bindings.LockPortRequest, emptypb.Empty]
	getUnlockedPorts            *connect.Client[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest, kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse]
}

// IsAvailable calls kurtosis_cloud.KurtosisCloudBackendServer.IsAvailable.
func (c *kurtosisCloudBackendServerClient) IsAvailable(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return c.isAvailable.CallUnary(ctx, req)
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

// RefreshDefaultPaymentMethod calls
// kurtosis_cloud.KurtosisCloudBackendServer.RefreshDefaultPaymentMethod.
func (c *kurtosisCloudBackendServerClient) RefreshDefaultPaymentMethod(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.refreshDefaultPaymentMethod.CallUnary(ctx, req)
}

// CancelPaymentSubscription calls
// kurtosis_cloud.KurtosisCloudBackendServer.CancelPaymentSubscription.
func (c *kurtosisCloudBackendServerClient) CancelPaymentSubscription(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.cancelPaymentSubscription.CallUnary(ctx, req)
}

// UpdateAddress calls kurtosis_cloud.KurtosisCloudBackendServer.UpdateAddress.
func (c *kurtosisCloudBackendServerClient) UpdateAddress(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs]) (*connect.Response[emptypb.Empty], error) {
	return c.updateAddress.CallUnary(ctx, req)
}

// GetInstances calls kurtosis_cloud.KurtosisCloudBackendServer.GetInstances.
func (c *kurtosisCloudBackendServerClient) GetInstances(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse], error) {
	return c.getInstances.CallUnary(ctx, req)
}

// DeleteInstance calls kurtosis_cloud.KurtosisCloudBackendServer.DeleteInstance.
func (c *kurtosisCloudBackendServerClient) DeleteInstance(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse], error) {
	return c.deleteInstance.CallUnary(ctx, req)
}

// ChangeActiveStatus calls kurtosis_cloud.KurtosisCloudBackendServer.ChangeActiveStatus.
func (c *kurtosisCloudBackendServerClient) ChangeActiveStatus(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest]) (*connect.Response[emptypb.Empty], error) {
	return c.changeActiveStatus.CallUnary(ctx, req)
}

// GetUser calls kurtosis_cloud.KurtosisCloudBackendServer.GetUser.
func (c *kurtosisCloudBackendServerClient) GetUser(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUserRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUserResponse], error) {
	return c.getUser.CallUnary(ctx, req)
}

// CheckPortAuthorization calls kurtosis_cloud.KurtosisCloudBackendServer.CheckPortAuthorization.
func (c *kurtosisCloudBackendServerClient) CheckPortAuthorization(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest]) (*connect.Response[emptypb.Empty], error) {
	return c.checkPortAuthorization.CallUnary(ctx, req)
}

// UnlockPort calls kurtosis_cloud.KurtosisCloudBackendServer.UnlockPort.
func (c *kurtosisCloudBackendServerClient) UnlockPort(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	return c.unlockPort.CallUnary(ctx, req)
}

// LockPort calls kurtosis_cloud.KurtosisCloudBackendServer.LockPort.
func (c *kurtosisCloudBackendServerClient) LockPort(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.LockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	return c.lockPort.CallUnary(ctx, req)
}

// GetUnlockedPorts calls kurtosis_cloud.KurtosisCloudBackendServer.GetUnlockedPorts.
func (c *kurtosisCloudBackendServerClient) GetUnlockedPorts(ctx context.Context, req *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse], error) {
	return c.getUnlockedPorts.CallUnary(ctx, req)
}

// KurtosisCloudBackendServerHandler is an implementation of the
// kurtosis_cloud.KurtosisCloudBackendServer service.
type KurtosisCloudBackendServerHandler interface {
	IsAvailable(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error)
	GetCloudInstanceConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetCloudInstanceConfigResponse], error)
	GetOrCreateApiKey(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateApiKeyResponse], error)
	GetOrCreateInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreateInstanceResponse], error)
	GetOrCreatePaymentConfig(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigArgs]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetOrCreatePaymentConfigResponse], error)
	RefreshDefaultPaymentMethod(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs]) (*connect.Response[emptypb.Empty], error)
	CancelPaymentSubscription(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs]) (*connect.Response[emptypb.Empty], error)
	UpdateAddress(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs]) (*connect.Response[emptypb.Empty], error)
	GetInstances(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse], error)
	DeleteInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse], error)
	ChangeActiveStatus(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest]) (*connect.Response[emptypb.Empty], error)
	GetUser(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUserRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUserResponse], error)
	CheckPortAuthorization(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest]) (*connect.Response[emptypb.Empty], error)
	UnlockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest]) (*connect.Response[emptypb.Empty], error)
	LockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.LockPortRequest]) (*connect.Response[emptypb.Empty], error)
	GetUnlockedPorts(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse], error)
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
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	kurtosisCloudBackendServerGetCloudInstanceConfigHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetCloudInstanceConfigProcedure,
		svc.GetCloudInstanceConfig,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
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
	kurtosisCloudBackendServerRefreshDefaultPaymentMethodHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerRefreshDefaultPaymentMethodProcedure,
		svc.RefreshDefaultPaymentMethod,
		opts...,
	)
	kurtosisCloudBackendServerCancelPaymentSubscriptionHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerCancelPaymentSubscriptionProcedure,
		svc.CancelPaymentSubscription,
		opts...,
	)
	kurtosisCloudBackendServerUpdateAddressHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerUpdateAddressProcedure,
		svc.UpdateAddress,
		opts...,
	)
	kurtosisCloudBackendServerGetInstancesHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetInstancesProcedure,
		svc.GetInstances,
		opts...,
	)
	kurtosisCloudBackendServerDeleteInstanceHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerDeleteInstanceProcedure,
		svc.DeleteInstance,
		opts...,
	)
	kurtosisCloudBackendServerChangeActiveStatusHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerChangeActiveStatusProcedure,
		svc.ChangeActiveStatus,
		opts...,
	)
	kurtosisCloudBackendServerGetUserHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetUserProcedure,
		svc.GetUser,
		opts...,
	)
	kurtosisCloudBackendServerCheckPortAuthorizationHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerCheckPortAuthorizationProcedure,
		svc.CheckPortAuthorization,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	kurtosisCloudBackendServerUnlockPortHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerUnlockPortProcedure,
		svc.UnlockPort,
		opts...,
	)
	kurtosisCloudBackendServerLockPortHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerLockPortProcedure,
		svc.LockPort,
		opts...,
	)
	kurtosisCloudBackendServerGetUnlockedPortsHandler := connect.NewUnaryHandler(
		KurtosisCloudBackendServerGetUnlockedPortsProcedure,
		svc.GetUnlockedPorts,
		opts...,
	)
	return "/kurtosis_cloud.KurtosisCloudBackendServer/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case KurtosisCloudBackendServerIsAvailableProcedure:
			kurtosisCloudBackendServerIsAvailableHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetCloudInstanceConfigProcedure:
			kurtosisCloudBackendServerGetCloudInstanceConfigHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreateApiKeyProcedure:
			kurtosisCloudBackendServerGetOrCreateApiKeyHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreateInstanceProcedure:
			kurtosisCloudBackendServerGetOrCreateInstanceHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetOrCreatePaymentConfigProcedure:
			kurtosisCloudBackendServerGetOrCreatePaymentConfigHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerRefreshDefaultPaymentMethodProcedure:
			kurtosisCloudBackendServerRefreshDefaultPaymentMethodHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerCancelPaymentSubscriptionProcedure:
			kurtosisCloudBackendServerCancelPaymentSubscriptionHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerUpdateAddressProcedure:
			kurtosisCloudBackendServerUpdateAddressHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetInstancesProcedure:
			kurtosisCloudBackendServerGetInstancesHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerDeleteInstanceProcedure:
			kurtosisCloudBackendServerDeleteInstanceHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerChangeActiveStatusProcedure:
			kurtosisCloudBackendServerChangeActiveStatusHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetUserProcedure:
			kurtosisCloudBackendServerGetUserHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerCheckPortAuthorizationProcedure:
			kurtosisCloudBackendServerCheckPortAuthorizationHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerUnlockPortProcedure:
			kurtosisCloudBackendServerUnlockPortHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerLockPortProcedure:
			kurtosisCloudBackendServerLockPortHandler.ServeHTTP(w, r)
		case KurtosisCloudBackendServerGetUnlockedPortsProcedure:
			kurtosisCloudBackendServerGetUnlockedPortsHandler.ServeHTTP(w, r)
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

func (UnimplementedKurtosisCloudBackendServerHandler) RefreshDefaultPaymentMethod(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.RefreshDefaultPaymentMethodArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.RefreshDefaultPaymentMethod is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) CancelPaymentSubscription(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CancelPaymentSubscriptionArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.CancelPaymentSubscription is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) UpdateAddress(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UpdateAddressArgs]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.UpdateAddress is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetInstances(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetInstancesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetInstances is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) DeleteInstance(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.DeleteInstanceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.DeleteInstance is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) ChangeActiveStatus(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.ChangeUserActiveRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.ChangeActiveStatus is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetUser(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUserRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetUser is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) CheckPortAuthorization(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.CheckPortAuthorizationRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.CheckPortAuthorization is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) UnlockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.UnlockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.UnlockPort is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) LockPort(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.LockPortRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.LockPort is not implemented"))
}

func (UnimplementedKurtosisCloudBackendServerHandler) GetUnlockedPorts(context.Context, *connect.Request[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsRequest]) (*connect.Response[kurtosis_backend_server_rpc_api_bindings.GetUnlockedPortsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("kurtosis_cloud.KurtosisCloudBackendServer.GetUnlockedPorts is not implemented"))
}
