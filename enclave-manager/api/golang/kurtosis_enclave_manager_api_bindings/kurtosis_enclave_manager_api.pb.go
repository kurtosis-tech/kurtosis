// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.3
// source: kurtosis_enclave_manager_api.proto

package kurtosis_enclave_manager_api_bindings

import (
	kurtosis_core_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	kurtosis_engine_rpc_api_bindings "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type HealthCheckResponse_ServingStatus int32

const (
	HealthCheckResponse_UNKNOWN         HealthCheckResponse_ServingStatus = 0
	HealthCheckResponse_SERVING         HealthCheckResponse_ServingStatus = 1
	HealthCheckResponse_NOT_SERVING     HealthCheckResponse_ServingStatus = 2
	HealthCheckResponse_SERVICE_UNKNOWN HealthCheckResponse_ServingStatus = 3 // Used only by the Watch method.
)

// Enum value maps for HealthCheckResponse_ServingStatus.
var (
	HealthCheckResponse_ServingStatus_name = map[int32]string{
		0: "UNKNOWN",
		1: "SERVING",
		2: "NOT_SERVING",
		3: "SERVICE_UNKNOWN",
	}
	HealthCheckResponse_ServingStatus_value = map[string]int32{
		"UNKNOWN":         0,
		"SERVING":         1,
		"NOT_SERVING":     2,
		"SERVICE_UNKNOWN": 3,
	}
)

func (x HealthCheckResponse_ServingStatus) Enum() *HealthCheckResponse_ServingStatus {
	p := new(HealthCheckResponse_ServingStatus)
	*p = x
	return p
}

func (x HealthCheckResponse_ServingStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (HealthCheckResponse_ServingStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_kurtosis_enclave_manager_api_proto_enumTypes[0].Descriptor()
}

func (HealthCheckResponse_ServingStatus) Type() protoreflect.EnumType {
	return &file_kurtosis_enclave_manager_api_proto_enumTypes[0]
}

func (x HealthCheckResponse_ServingStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use HealthCheckResponse_ServingStatus.Descriptor instead.
func (HealthCheckResponse_ServingStatus) EnumDescriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{1, 0}
}

type HealthCheckRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service string `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
}

func (x *HealthCheckRequest) Reset() {
	*x = HealthCheckRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthCheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckRequest) ProtoMessage() {}

func (x *HealthCheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckRequest.ProtoReflect.Descriptor instead.
func (*HealthCheckRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{0}
}

func (x *HealthCheckRequest) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

type HealthCheckResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status HealthCheckResponse_ServingStatus `protobuf:"varint,1,opt,name=status,proto3,enum=kurtosis_enclave_manager.HealthCheckResponse_ServingStatus" json:"status,omitempty"`
}

func (x *HealthCheckResponse) Reset() {
	*x = HealthCheckResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HealthCheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckResponse) ProtoMessage() {}

func (x *HealthCheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HealthCheckResponse.ProtoReflect.Descriptor instead.
func (*HealthCheckResponse) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{1}
}

func (x *HealthCheckResponse) GetStatus() HealthCheckResponse_ServingStatus {
	if x != nil {
		return x.Status
	}
	return HealthCheckResponse_UNKNOWN
}

type GetServicesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApicIpAddress string `protobuf:"bytes,1,opt,name=apic_ip_address,json=apicIpAddress,proto3" json:"apic_ip_address,omitempty"`
	ApicPort      int32  `protobuf:"varint,2,opt,name=apic_port,json=apicPort,proto3" json:"apic_port,omitempty"`
}

func (x *GetServicesRequest) Reset() {
	*x = GetServicesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetServicesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetServicesRequest) ProtoMessage() {}

func (x *GetServicesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetServicesRequest.ProtoReflect.Descriptor instead.
func (*GetServicesRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{2}
}

func (x *GetServicesRequest) GetApicIpAddress() string {
	if x != nil {
		return x.ApicIpAddress
	}
	return ""
}

func (x *GetServicesRequest) GetApicPort() int32 {
	if x != nil {
		return x.ApicPort
	}
	return 0
}

type GetListFilesArtifactNamesAndUuidsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApicIpAddress string `protobuf:"bytes,1,opt,name=apic_ip_address,json=apicIpAddress,proto3" json:"apic_ip_address,omitempty"`
	ApicPort      int32  `protobuf:"varint,2,opt,name=apic_port,json=apicPort,proto3" json:"apic_port,omitempty"`
}

func (x *GetListFilesArtifactNamesAndUuidsRequest) Reset() {
	*x = GetListFilesArtifactNamesAndUuidsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetListFilesArtifactNamesAndUuidsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetListFilesArtifactNamesAndUuidsRequest) ProtoMessage() {}

func (x *GetListFilesArtifactNamesAndUuidsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetListFilesArtifactNamesAndUuidsRequest.ProtoReflect.Descriptor instead.
func (*GetListFilesArtifactNamesAndUuidsRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{3}
}

func (x *GetListFilesArtifactNamesAndUuidsRequest) GetApicIpAddress() string {
	if x != nil {
		return x.ApicIpAddress
	}
	return ""
}

func (x *GetListFilesArtifactNamesAndUuidsRequest) GetApicPort() int32 {
	if x != nil {
		return x.ApicPort
	}
	return 0
}

type RunStarlarkPackageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApicIpAddress          string                                                 `protobuf:"bytes,1,opt,name=apic_ip_address,json=apicIpAddress,proto3" json:"apic_ip_address,omitempty"`
	ApicPort               int32                                                  `protobuf:"varint,2,opt,name=apic_port,json=apicPort,proto3" json:"apic_port,omitempty"`
	RunStarlarkPackageArgs *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs `protobuf:"bytes,3,opt,name=RunStarlarkPackageArgs,proto3" json:"RunStarlarkPackageArgs,omitempty"`
}

func (x *RunStarlarkPackageRequest) Reset() {
	*x = RunStarlarkPackageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunStarlarkPackageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunStarlarkPackageRequest) ProtoMessage() {}

func (x *RunStarlarkPackageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunStarlarkPackageRequest.ProtoReflect.Descriptor instead.
func (*RunStarlarkPackageRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{4}
}

func (x *RunStarlarkPackageRequest) GetApicIpAddress() string {
	if x != nil {
		return x.ApicIpAddress
	}
	return ""
}

func (x *RunStarlarkPackageRequest) GetApicPort() int32 {
	if x != nil {
		return x.ApicPort
	}
	return 0
}

func (x *RunStarlarkPackageRequest) GetRunStarlarkPackageArgs() *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs {
	if x != nil {
		return x.RunStarlarkPackageArgs
	}
	return nil
}

type InspectFilesArtifactContentsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApicIpAddress    string                                                   `protobuf:"bytes,1,opt,name=apic_ip_address,json=apicIpAddress,proto3" json:"apic_ip_address,omitempty"`
	ApicPort         int32                                                    `protobuf:"varint,2,opt,name=apic_port,json=apicPort,proto3" json:"apic_port,omitempty"`
	FileNamesAndUuid *kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid `protobuf:"bytes,3,opt,name=file_names_and_uuid,json=fileNamesAndUuid,proto3" json:"file_names_and_uuid,omitempty"`
}

func (x *InspectFilesArtifactContentsRequest) Reset() {
	*x = InspectFilesArtifactContentsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InspectFilesArtifactContentsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InspectFilesArtifactContentsRequest) ProtoMessage() {}

func (x *InspectFilesArtifactContentsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InspectFilesArtifactContentsRequest.ProtoReflect.Descriptor instead.
func (*InspectFilesArtifactContentsRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{5}
}

func (x *InspectFilesArtifactContentsRequest) GetApicIpAddress() string {
	if x != nil {
		return x.ApicIpAddress
	}
	return ""
}

func (x *InspectFilesArtifactContentsRequest) GetApicPort() int32 {
	if x != nil {
		return x.ApicPort
	}
	return 0
}

func (x *InspectFilesArtifactContentsRequest) GetFileNamesAndUuid() *kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid {
	if x != nil {
		return x.FileNamesAndUuid
	}
	return nil
}

type GetStarlarkRunRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ApicIpAddress string `protobuf:"bytes,1,opt,name=apic_ip_address,json=apicIpAddress,proto3" json:"apic_ip_address,omitempty"`
	ApicPort      int32  `protobuf:"varint,2,opt,name=apic_port,json=apicPort,proto3" json:"apic_port,omitempty"`
}

func (x *GetStarlarkRunRequest) Reset() {
	*x = GetStarlarkRunRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStarlarkRunRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStarlarkRunRequest) ProtoMessage() {}

func (x *GetStarlarkRunRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kurtosis_enclave_manager_api_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStarlarkRunRequest.ProtoReflect.Descriptor instead.
func (*GetStarlarkRunRequest) Descriptor() ([]byte, []int) {
	return file_kurtosis_enclave_manager_api_proto_rawDescGZIP(), []int{6}
}

func (x *GetStarlarkRunRequest) GetApicIpAddress() string {
	if x != nil {
		return x.ApicIpAddress
	}
	return ""
}

func (x *GetStarlarkRunRequest) GetApicPort() int32 {
	if x != nil {
		return x.ApicPort
	}
	return 0
}

var File_kurtosis_enclave_manager_api_proto protoreflect.FileDescriptor

var file_kurtosis_enclave_manager_api_proto_rawDesc = []byte{
	0x0a, 0x22, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61,
	0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65,
	0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x1a, 0x1b,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x2e,
	0x0a, 0x12, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x22, 0xbb,
	0x01, 0x0a, 0x13, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x53, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x3b, 0x2e, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69,
	0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65,
	0x72, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x4f, 0x0a, 0x0d, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07,
	0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x45, 0x52,
	0x56, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x4e, 0x4f, 0x54, 0x5f, 0x53, 0x45,
	0x52, 0x56, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x45, 0x52, 0x56, 0x49,
	0x43, 0x45, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x03, 0x22, 0x59, 0x0a, 0x12,
	0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x70, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x61, 0x70, 0x69,
	0x63, 0x49, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x70,
	0x69, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61,
	0x70, 0x69, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x6f, 0x0a, 0x28, 0x47, 0x65, 0x74, 0x4c, 0x69,
	0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x4e,
	0x61, 0x6d, 0x65, 0x73, 0x41, 0x6e, 0x64, 0x55, 0x75, 0x69, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x70, 0x5f, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x61, 0x70,
	0x69, 0x63, 0x49, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x61,
	0x70, 0x69, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x61, 0x70, 0x69, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x22, 0xc3, 0x01, 0x0a, 0x19, 0x52, 0x75, 0x6e,
	0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x69,
	0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x61, 0x70, 0x69, 0x63, 0x49, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x61, 0x70, 0x69, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x61, 0x0a, 0x16, 0x52,
	0x75, 0x6e, 0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x41, 0x72, 0x67, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x61, 0x70,
	0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e,
	0x52, 0x75, 0x6e, 0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61,
	0x67, 0x65, 0x41, 0x72, 0x67, 0x73, 0x52, 0x16, 0x52, 0x75, 0x6e, 0x53, 0x74, 0x61, 0x72, 0x6c,
	0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x41, 0x72, 0x67, 0x73, 0x22, 0xc6,
	0x01, 0x0a, 0x23, 0x49, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41,
	0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x69,
	0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x61, 0x70, 0x69, 0x63, 0x49, 0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x61, 0x70, 0x69, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x12, 0x5a, 0x0a, 0x13, 0x66,
	0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x5f, 0x61, 0x6e, 0x64, 0x5f, 0x75, 0x75,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x41, 0x6e,
	0x64, 0x55, 0x75, 0x69, 0x64, 0x52, 0x10, 0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x41, 0x6e, 0x64, 0x55, 0x75, 0x69, 0x64, 0x22, 0x5c, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x26, 0x0a, 0x0f, 0x61, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x61, 0x70, 0x69, 0x63, 0x49,
	0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x70, 0x69, 0x63,
	0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x61, 0x70, 0x69,
	0x63, 0x50, 0x6f, 0x72, 0x74, 0x32, 0xcb, 0x09, 0x0a, 0x1c, 0x4b, 0x75, 0x72, 0x74, 0x6f, 0x73,
	0x69, 0x73, 0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x64, 0x0a, 0x05, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12,
	0x2c, 0x2e, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61,
	0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e,
	0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65,
	0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x1f, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x5f, 0x61, 0x70, 0x69,
	0x2e, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x65, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x12, 0x2c, 0x2e, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73,
	0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69,
	0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x58, 0x0a,
	0x0e, 0x47, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4c, 0x6f, 0x67, 0x73, 0x12,
	0x1e, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4c, 0x6f, 0x67, 0x73, 0x41, 0x72, 0x67, 0x73, 0x1a,
	0x22, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0xa1, 0x01, 0x0a, 0x1e, 0x4c, 0x69, 0x73, 0x74,
	0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x4e, 0x61, 0x6d,
	0x65, 0x73, 0x41, 0x6e, 0x64, 0x55, 0x75, 0x69, 0x64, 0x73, 0x12, 0x42, 0x2e, 0x6b, 0x75, 0x72,
	0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61,
	0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c,
	0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x41,
	0x6e, 0x64, 0x55, 0x75, 0x69, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x39,
	0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61,
	0x70, 0x69, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69,
	0x66, 0x61, 0x63, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x41, 0x6e, 0x64, 0x55, 0x75, 0x69, 0x64,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x79, 0x0a, 0x12, 0x52,
	0x75, 0x6e, 0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67,
	0x65, 0x12, 0x33, 0x2e, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63,
	0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x52, 0x75, 0x6e,
	0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x50, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x74, 0x61, 0x72, 0x6c,
	0x61, 0x72, 0x6b, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4c, 0x69,
	0x6e, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x53, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x12, 0x1d, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65,
	0x5f, 0x61, 0x70, 0x69, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x63, 0x6c, 0x61,
	0x76, 0x65, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x21, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x5f,
	0x61, 0x70, 0x69, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x6f, 0x0a, 0x15, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69,
	0x66, 0x61, 0x63, 0x74, 0x12, 0x2c, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x41, 0x72,
	0x67, 0x73, 0x1a, 0x24, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x65, 0x64, 0x44,
	0x61, 0x74, 0x61, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x22, 0x00, 0x30, 0x01, 0x12, 0x98, 0x01, 0x0a,
	0x1c, 0x49, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74,
	0x69, 0x66, 0x61, 0x63, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x3d, 0x2e,
	0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65,
	0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x49, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74,
	0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x37, 0x2e, 0x61,
	0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69,
	0x2e, 0x49, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x41, 0x72, 0x74,
	0x69, 0x66, 0x61, 0x63, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4a, 0x0a, 0x0e, 0x44, 0x65, 0x73, 0x74, 0x72,
	0x6f, 0x79, 0x45, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x12, 0x1e, 0x2e, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x44, 0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x45, 0x6e,
	0x63, 0x6c, 0x61, 0x76, 0x65, 0x41, 0x72, 0x67, 0x73, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x12, 0x6e, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x6c, 0x61,
	0x72, 0x6b, 0x52, 0x75, 0x6e, 0x12, 0x2f, 0x2e, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73,
	0x5f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x52, 0x75, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x61, 0x70, 0x69, 0x5f, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x72, 0x6c, 0x61, 0x72, 0x6b, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x64, 0x5a, 0x62, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x2d, 0x74, 0x65, 0x63, 0x68, 0x2f,
	0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x2f, 0x65, 0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65,
	0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x6f, 0x6c,
	0x61, 0x6e, 0x67, 0x2f, 0x6b, 0x75, 0x72, 0x74, 0x6f, 0x73, 0x69, 0x73, 0x5f, 0x65, 0x6e, 0x63,
	0x6c, 0x61, 0x76, 0x65, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x5f, 0x61, 0x70, 0x69,
	0x5f, 0x62, 0x69, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_kurtosis_enclave_manager_api_proto_rawDescOnce sync.Once
	file_kurtosis_enclave_manager_api_proto_rawDescData = file_kurtosis_enclave_manager_api_proto_rawDesc
)

func file_kurtosis_enclave_manager_api_proto_rawDescGZIP() []byte {
	file_kurtosis_enclave_manager_api_proto_rawDescOnce.Do(func() {
		file_kurtosis_enclave_manager_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_kurtosis_enclave_manager_api_proto_rawDescData)
	})
	return file_kurtosis_enclave_manager_api_proto_rawDescData
}

var file_kurtosis_enclave_manager_api_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_kurtosis_enclave_manager_api_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_kurtosis_enclave_manager_api_proto_goTypes = []interface{}{
	(HealthCheckResponse_ServingStatus)(0),                                        // 0: kurtosis_enclave_manager.HealthCheckResponse.ServingStatus
	(*HealthCheckRequest)(nil),                                                    // 1: kurtosis_enclave_manager.HealthCheckRequest
	(*HealthCheckResponse)(nil),                                                   // 2: kurtosis_enclave_manager.HealthCheckResponse
	(*GetServicesRequest)(nil),                                                    // 3: kurtosis_enclave_manager.GetServicesRequest
	(*GetListFilesArtifactNamesAndUuidsRequest)(nil),                              // 4: kurtosis_enclave_manager.GetListFilesArtifactNamesAndUuidsRequest
	(*RunStarlarkPackageRequest)(nil),                                             // 5: kurtosis_enclave_manager.RunStarlarkPackageRequest
	(*InspectFilesArtifactContentsRequest)(nil),                                   // 6: kurtosis_enclave_manager.InspectFilesArtifactContentsRequest
	(*GetStarlarkRunRequest)(nil),                                                 // 7: kurtosis_enclave_manager.GetStarlarkRunRequest
	(*kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs)(nil),                 // 8: api_container_api.RunStarlarkPackageArgs
	(*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid)(nil),               // 9: api_container_api.FilesArtifactNameAndUuid
	(*emptypb.Empty)(nil),                                                         // 10: google.protobuf.Empty
	(*kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs)(nil),                   // 11: engine_api.GetServiceLogsArgs
	(*kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs)(nil),                    // 12: engine_api.CreateEnclaveArgs
	(*kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs)(nil),              // 13: api_container_api.DownloadFilesArtifactArgs
	(*kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs)(nil),                   // 14: engine_api.DestroyEnclaveArgs
	(*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse)(nil),                  // 15: engine_api.GetEnclavesResponse
	(*kurtosis_core_rpc_api_bindings.GetServicesResponse)(nil),                    // 16: api_container_api.GetServicesResponse
	(*kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse)(nil),               // 17: engine_api.GetServiceLogsResponse
	(*kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse)(nil), // 18: api_container_api.ListFilesArtifactNamesAndUuidsResponse
	(*kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)(nil),                // 19: api_container_api.StarlarkRunResponseLine
	(*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse)(nil),                // 20: engine_api.CreateEnclaveResponse
	(*kurtosis_core_rpc_api_bindings.StreamedDataChunk)(nil),                      // 21: api_container_api.StreamedDataChunk
	(*kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse)(nil),   // 22: api_container_api.InspectFilesArtifactContentsResponse
	(*kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse)(nil),                 // 23: api_container_api.GetStarlarkRunResponse
}
var file_kurtosis_enclave_manager_api_proto_depIdxs = []int32{
	0,  // 0: kurtosis_enclave_manager.HealthCheckResponse.status:type_name -> kurtosis_enclave_manager.HealthCheckResponse.ServingStatus
	8,  // 1: kurtosis_enclave_manager.RunStarlarkPackageRequest.RunStarlarkPackageArgs:type_name -> api_container_api.RunStarlarkPackageArgs
	9,  // 2: kurtosis_enclave_manager.InspectFilesArtifactContentsRequest.file_names_and_uuid:type_name -> api_container_api.FilesArtifactNameAndUuid
	1,  // 3: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.Check:input_type -> kurtosis_enclave_manager.HealthCheckRequest
	10, // 4: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetEnclaves:input_type -> google.protobuf.Empty
	3,  // 5: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServices:input_type -> kurtosis_enclave_manager.GetServicesRequest
	11, // 6: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServiceLogs:input_type -> engine_api.GetServiceLogsArgs
	4,  // 7: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.ListFilesArtifactNamesAndUuids:input_type -> kurtosis_enclave_manager.GetListFilesArtifactNamesAndUuidsRequest
	5,  // 8: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.RunStarlarkPackage:input_type -> kurtosis_enclave_manager.RunStarlarkPackageRequest
	12, // 9: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.CreateEnclave:input_type -> engine_api.CreateEnclaveArgs
	13, // 10: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DownloadFilesArtifact:input_type -> api_container_api.DownloadFilesArtifactArgs
	6,  // 11: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.InspectFilesArtifactContents:input_type -> kurtosis_enclave_manager.InspectFilesArtifactContentsRequest
	14, // 12: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DestroyEnclave:input_type -> engine_api.DestroyEnclaveArgs
	7,  // 13: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetStarlarkRun:input_type -> kurtosis_enclave_manager.GetStarlarkRunRequest
	2,  // 14: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.Check:output_type -> kurtosis_enclave_manager.HealthCheckResponse
	15, // 15: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetEnclaves:output_type -> engine_api.GetEnclavesResponse
	16, // 16: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServices:output_type -> api_container_api.GetServicesResponse
	17, // 17: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServiceLogs:output_type -> engine_api.GetServiceLogsResponse
	18, // 18: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.ListFilesArtifactNamesAndUuids:output_type -> api_container_api.ListFilesArtifactNamesAndUuidsResponse
	19, // 19: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.RunStarlarkPackage:output_type -> api_container_api.StarlarkRunResponseLine
	20, // 20: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.CreateEnclave:output_type -> engine_api.CreateEnclaveResponse
	21, // 21: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DownloadFilesArtifact:output_type -> api_container_api.StreamedDataChunk
	22, // 22: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.InspectFilesArtifactContents:output_type -> api_container_api.InspectFilesArtifactContentsResponse
	10, // 23: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.DestroyEnclave:output_type -> google.protobuf.Empty
	23, // 24: kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetStarlarkRun:output_type -> api_container_api.GetStarlarkRunResponse
	14, // [14:25] is the sub-list for method output_type
	3,  // [3:14] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_kurtosis_enclave_manager_api_proto_init() }
func file_kurtosis_enclave_manager_api_proto_init() {
	if File_kurtosis_enclave_manager_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kurtosis_enclave_manager_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthCheckRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HealthCheckResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetServicesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetListFilesArtifactNamesAndUuidsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunStarlarkPackageRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InspectFilesArtifactContentsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_kurtosis_enclave_manager_api_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStarlarkRunRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kurtosis_enclave_manager_api_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kurtosis_enclave_manager_api_proto_goTypes,
		DependencyIndexes: file_kurtosis_enclave_manager_api_proto_depIdxs,
		EnumInfos:         file_kurtosis_enclave_manager_api_proto_enumTypes,
		MessageInfos:      file_kurtosis_enclave_manager_api_proto_msgTypes,
	}.Build()
	File_kurtosis_enclave_manager_api_proto = out.File
	file_kurtosis_enclave_manager_api_proto_rawDesc = nil
	file_kurtosis_enclave_manager_api_proto_goTypes = nil
	file_kurtosis_enclave_manager_api_proto_depIdxs = nil
}
