// GENERATED CODE -- DO NOT EDIT!

// package: module_api
// file: executable_module_service.proto

import * as executable_module_service_pb from "./executable_module_service_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";

interface IExecutableModuleServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  isAvailable: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
  execute: grpc.MethodDefinition<executable_module_service_pb.ExecuteArgs, executable_module_service_pb.ExecuteResponse>;
}

export const ExecutableModuleServiceService: IExecutableModuleServiceService;

export interface IExecutableModuleServiceServer extends grpc.UntypedServiceImplementation {
  isAvailable: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
  execute: grpc.handleUnaryCall<executable_module_service_pb.ExecuteArgs, executable_module_service_pb.ExecuteResponse>;
}

export class ExecutableModuleServiceClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  isAvailable(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  isAvailable(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  isAvailable(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  execute(argument: executable_module_service_pb.ExecuteArgs, callback: grpc.requestCallback<executable_module_service_pb.ExecuteResponse>): grpc.ClientUnaryCall;
  execute(argument: executable_module_service_pb.ExecuteArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<executable_module_service_pb.ExecuteResponse>): grpc.ClientUnaryCall;
  execute(argument: executable_module_service_pb.ExecuteArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<executable_module_service_pb.ExecuteResponse>): grpc.ClientUnaryCall;
}
