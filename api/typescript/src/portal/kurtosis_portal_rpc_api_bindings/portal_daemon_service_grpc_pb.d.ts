// GENERATED CODE -- DO NOT EDIT!

// package: portal_daemon_api
// file: portal_daemon_service.proto

import * as portal_daemon_service_pb from "./portal_daemon_service_pb";
import * as grpc from "@grpc/grpc-js";

interface IKurtosisPortalDaemonService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  ping: grpc.MethodDefinition<portal_daemon_service_pb.PortalPing, portal_daemon_service_pb.PortalPong>;
  createUserServicePortForward: grpc.MethodDefinition<portal_daemon_service_pb.CreateUserServicePortForwardArgs, portal_daemon_service_pb.CreateUserServicePortForwardResponse>;
  removeUserServicePortForward: grpc.MethodDefinition<portal_daemon_service_pb.EnclaveServicePortId, portal_daemon_service_pb.RemoveUserServicePortForwardResponse>;
}

export const KurtosisPortalDaemonService: IKurtosisPortalDaemonService;

export interface IKurtosisPortalDaemonServer extends grpc.UntypedServiceImplementation {
  ping: grpc.handleUnaryCall<portal_daemon_service_pb.PortalPing, portal_daemon_service_pb.PortalPong>;
  createUserServicePortForward: grpc.handleUnaryCall<portal_daemon_service_pb.CreateUserServicePortForwardArgs, portal_daemon_service_pb.CreateUserServicePortForwardResponse>;
  removeUserServicePortForward: grpc.handleUnaryCall<portal_daemon_service_pb.EnclaveServicePortId, portal_daemon_service_pb.RemoveUserServicePortForwardResponse>;
}

export class KurtosisPortalDaemonClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  ping(argument: portal_daemon_service_pb.PortalPing, callback: grpc.requestCallback<portal_daemon_service_pb.PortalPong>): grpc.ClientUnaryCall;
  ping(argument: portal_daemon_service_pb.PortalPing, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.PortalPong>): grpc.ClientUnaryCall;
  ping(argument: portal_daemon_service_pb.PortalPing, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.PortalPong>): grpc.ClientUnaryCall;
  createUserServicePortForward(argument: portal_daemon_service_pb.CreateUserServicePortForwardArgs, callback: grpc.requestCallback<portal_daemon_service_pb.CreateUserServicePortForwardResponse>): grpc.ClientUnaryCall;
  createUserServicePortForward(argument: portal_daemon_service_pb.CreateUserServicePortForwardArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.CreateUserServicePortForwardResponse>): grpc.ClientUnaryCall;
  createUserServicePortForward(argument: portal_daemon_service_pb.CreateUserServicePortForwardArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.CreateUserServicePortForwardResponse>): grpc.ClientUnaryCall;
  removeUserServicePortForward(argument: portal_daemon_service_pb.EnclaveServicePortId, callback: grpc.requestCallback<portal_daemon_service_pb.RemoveUserServicePortForwardResponse>): grpc.ClientUnaryCall;
  removeUserServicePortForward(argument: portal_daemon_service_pb.EnclaveServicePortId, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.RemoveUserServicePortForwardResponse>): grpc.ClientUnaryCall;
  removeUserServicePortForward(argument: portal_daemon_service_pb.EnclaveServicePortId, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<portal_daemon_service_pb.RemoveUserServicePortForwardResponse>): grpc.ClientUnaryCall;
}
