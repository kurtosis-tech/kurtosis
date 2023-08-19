// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.20.3
// source: kurtosis_enclave_manager_api.proto

package kurtosis_enclave_manager_api_bindings

import (
	context "context"
	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	kurtosis_engine_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	KurtosisEnclaveManagerServer_Check_FullMethodName                          = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/Check"
	KurtosisEnclaveManagerServer_GetEnclaves_FullMethodName                    = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetEnclaves"
	KurtosisEnclaveManagerServer_GetServices_FullMethodName                    = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetServices"
	KurtosisEnclaveManagerServer_CreateEnclave_FullMethodName                  = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/CreateEnclave"
	KurtosisEnclaveManagerServer_GetServiceLogs_FullMethodName                 = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/GetServiceLogs"
	KurtosisEnclaveManagerServer_RunStarlarkPackage_FullMethodName             = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/RunStarlarkPackage"
	KurtosisEnclaveManagerServer_ListFilesArtifactNamesAndUuids_FullMethodName = "/kurtosis_enclave_manager.KurtosisEnclaveManagerServer/ListFilesArtifactNamesAndUuids"
)

// KurtosisEnclaveManagerServerClient is the client API for KurtosisEnclaveManagerServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KurtosisEnclaveManagerServerClient interface {
	Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error)
	GetEnclaves(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error)
	GetServices(ctx context.Context, in *GetServicesRequest, opts ...grpc.CallOption) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error)
	CreateEnclave(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetServiceLogs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RunStarlarkPackage(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ListFilesArtifactNamesAndUuids(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type kurtosisEnclaveManagerServerClient struct {
	cc grpc.ClientConnInterface
}

func NewKurtosisEnclaveManagerServerClient(cc grpc.ClientConnInterface) KurtosisEnclaveManagerServerClient {
	return &kurtosisEnclaveManagerServerClient{cc}
}

func (c *kurtosisEnclaveManagerServerClient) Check(ctx context.Context, in *HealthCheckRequest, opts ...grpc.CallOption) (*HealthCheckResponse, error) {
	out := new(HealthCheckResponse)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_Check_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) GetEnclaves(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	out := new(kurtosis_engine_rpc_api_bindings.GetEnclavesResponse)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_GetEnclaves_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) GetServices(ctx context.Context, in *GetServicesRequest, opts ...grpc.CallOption) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	out := new(kurtosis_core_rpc_api_bindings.GetServicesResponse)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_GetServices_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) CreateEnclave(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_CreateEnclave_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) GetServiceLogs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_GetServiceLogs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) RunStarlarkPackage(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_RunStarlarkPackage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kurtosisEnclaveManagerServerClient) ListFilesArtifactNamesAndUuids(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, KurtosisEnclaveManagerServer_ListFilesArtifactNamesAndUuids_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KurtosisEnclaveManagerServerServer is the server API for KurtosisEnclaveManagerServer service.
// All implementations should embed UnimplementedKurtosisEnclaveManagerServerServer
// for forward compatibility
type KurtosisEnclaveManagerServerServer interface {
	Check(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error)
	GetEnclaves(context.Context, *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error)
	GetServices(context.Context, *GetServicesRequest) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error)
	CreateEnclave(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	GetServiceLogs(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	RunStarlarkPackage(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	ListFilesArtifactNamesAndUuids(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
}

// UnimplementedKurtosisEnclaveManagerServerServer should be embedded to have forward compatible implementations.
type UnimplementedKurtosisEnclaveManagerServerServer struct {
}

func (UnimplementedKurtosisEnclaveManagerServerServer) Check(context.Context, *HealthCheckRequest) (*HealthCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) GetEnclaves(context.Context, *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEnclaves not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) GetServices(context.Context, *GetServicesRequest) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServices not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) CreateEnclave(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateEnclave not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) GetServiceLogs(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetServiceLogs not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) RunStarlarkPackage(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RunStarlarkPackage not implemented")
}
func (UnimplementedKurtosisEnclaveManagerServerServer) ListFilesArtifactNamesAndUuids(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListFilesArtifactNamesAndUuids not implemented")
}

// UnsafeKurtosisEnclaveManagerServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KurtosisEnclaveManagerServerServer will
// result in compilation errors.
type UnsafeKurtosisEnclaveManagerServerServer interface {
	mustEmbedUnimplementedKurtosisEnclaveManagerServerServer()
}

func RegisterKurtosisEnclaveManagerServerServer(s grpc.ServiceRegistrar, srv KurtosisEnclaveManagerServerServer) {
	s.RegisterService(&KurtosisEnclaveManagerServer_ServiceDesc, srv)
}

func _KurtosisEnclaveManagerServer_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_Check_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).Check(ctx, req.(*HealthCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_GetEnclaves_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).GetEnclaves(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_GetEnclaves_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).GetEnclaves(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_GetServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetServicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).GetServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_GetServices_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).GetServices(ctx, req.(*GetServicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_CreateEnclave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).CreateEnclave(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_CreateEnclave_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).CreateEnclave(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_GetServiceLogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).GetServiceLogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_GetServiceLogs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).GetServiceLogs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_RunStarlarkPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).RunStarlarkPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_RunStarlarkPackage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).RunStarlarkPackage(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _KurtosisEnclaveManagerServer_ListFilesArtifactNamesAndUuids_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KurtosisEnclaveManagerServerServer).ListFilesArtifactNamesAndUuids(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: KurtosisEnclaveManagerServer_ListFilesArtifactNamesAndUuids_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KurtosisEnclaveManagerServerServer).ListFilesArtifactNamesAndUuids(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// KurtosisEnclaveManagerServer_ServiceDesc is the grpc.ServiceDesc for KurtosisEnclaveManagerServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KurtosisEnclaveManagerServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kurtosis_enclave_manager.KurtosisEnclaveManagerServer",
	HandlerType: (*KurtosisEnclaveManagerServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _KurtosisEnclaveManagerServer_Check_Handler,
		},
		{
			MethodName: "GetEnclaves",
			Handler:    _KurtosisEnclaveManagerServer_GetEnclaves_Handler,
		},
		{
			MethodName: "GetServices",
			Handler:    _KurtosisEnclaveManagerServer_GetServices_Handler,
		},
		{
			MethodName: "CreateEnclave",
			Handler:    _KurtosisEnclaveManagerServer_CreateEnclave_Handler,
		},
		{
			MethodName: "GetServiceLogs",
			Handler:    _KurtosisEnclaveManagerServer_GetServiceLogs_Handler,
		},
		{
			MethodName: "RunStarlarkPackage",
			Handler:    _KurtosisEnclaveManagerServer_RunStarlarkPackage_Handler,
		},
		{
			MethodName: "ListFilesArtifactNamesAndUuids",
			Handler:    _KurtosisEnclaveManagerServer_ListFilesArtifactNamesAndUuids_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kurtosis_enclave_manager_api.proto",
}
