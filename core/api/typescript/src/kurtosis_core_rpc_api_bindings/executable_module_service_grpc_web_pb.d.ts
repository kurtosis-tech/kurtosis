import * as grpcWeb from 'grpc-web';

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';
import * as executable_module_service_pb from './executable_module_service_pb';


export class ExecutableModuleServiceClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  isAvailable(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  execute(
    request: executable_module_service_pb.ExecuteArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: executable_module_service_pb.ExecuteResponse) => void
  ): grpcWeb.ClientReadableStream<executable_module_service_pb.ExecuteResponse>;

}

export class ExecutableModuleServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  isAvailable(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  execute(
    request: executable_module_service_pb.ExecuteArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<executable_module_service_pb.ExecuteResponse>;

}

