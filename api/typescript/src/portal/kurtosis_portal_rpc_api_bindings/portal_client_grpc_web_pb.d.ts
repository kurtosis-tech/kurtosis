import * as grpcWeb from 'grpc-web';

import * as portal_client_pb from './portal_client_pb'; // proto import: "portal_client.proto"


export class KurtosisPortalDaemonClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  ping(
    request: portal_client_pb.PortalPing,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: portal_client_pb.PortalPong) => void
  ): grpcWeb.ClientReadableStream<portal_client_pb.PortalPong>;

  forwardUserServicePort(
    request: portal_client_pb.ForwardUserServicePortArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: portal_client_pb.ForwardUserServicePortResponse) => void
  ): grpcWeb.ClientReadableStream<portal_client_pb.ForwardUserServicePortResponse>;

}

export class KurtosisPortalDaemonPromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  ping(
    request: portal_client_pb.PortalPing,
    metadata?: grpcWeb.Metadata
  ): Promise<portal_client_pb.PortalPong>;

  forwardUserServicePort(
    request: portal_client_pb.ForwardUserServicePortArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<portal_client_pb.ForwardUserServicePortResponse>;

}

