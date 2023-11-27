import * as grpcWeb from 'grpc-web';

import * as portal_daemon_service_pb from './portal_daemon_service_pb'; // proto import: "portal_daemon_service.proto"


export class KurtosisPortalDaemonClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  ping(
    request: portal_daemon_service_pb.PortalPing,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: portal_daemon_service_pb.PortalPong) => void
  ): grpcWeb.ClientReadableStream<portal_daemon_service_pb.PortalPong>;

  createUserServicePortForward(
    request: portal_daemon_service_pb.CreateUserServicePortForwardArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: portal_daemon_service_pb.CreateUserServicePortForwardResponse) => void
  ): grpcWeb.ClientReadableStream<portal_daemon_service_pb.CreateUserServicePortForwardResponse>;

  removeUserServicePortForward(
    request: portal_daemon_service_pb.EnclaveServicePortId,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: portal_daemon_service_pb.RemoveUserServicePortForwardResponse) => void
  ): grpcWeb.ClientReadableStream<portal_daemon_service_pb.RemoveUserServicePortForwardResponse>;

}

export class KurtosisPortalDaemonPromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  ping(
    request: portal_daemon_service_pb.PortalPing,
    metadata?: grpcWeb.Metadata
  ): Promise<portal_daemon_service_pb.PortalPong>;

  createUserServicePortForward(
    request: portal_daemon_service_pb.CreateUserServicePortForwardArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<portal_daemon_service_pb.CreateUserServicePortForwardResponse>;

  removeUserServicePortForward(
    request: portal_daemon_service_pb.EnclaveServicePortId,
    metadata?: grpcWeb.Metadata
  ): Promise<portal_daemon_service_pb.RemoveUserServicePortForwardResponse>;

}

