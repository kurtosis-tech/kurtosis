import * as grpcWeb from 'grpc-web';

import * as engine_service_pb from './engine_service_pb';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class EngineServiceClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  getEngineInfo(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: engine_service_pb.GetEngineInfoResponse) => void
  ): grpcWeb.ClientReadableStream<engine_service_pb.GetEngineInfoResponse>;

  createEnclave(
    request: engine_service_pb.CreateEnclaveArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: engine_service_pb.CreateEnclaveResponse) => void
  ): grpcWeb.ClientReadableStream<engine_service_pb.CreateEnclaveResponse>;

  getEnclaves(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: engine_service_pb.GetEnclavesResponse) => void
  ): grpcWeb.ClientReadableStream<engine_service_pb.GetEnclavesResponse>;

  getExistingAndHistoricalEnclaveIdentifiers(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse) => void
  ): grpcWeb.ClientReadableStream<engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse>;

  stopEnclave(
    request: engine_service_pb.StopEnclaveArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  destroyEnclave(
    request: engine_service_pb.DestroyEnclaveArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  clean(
    request: engine_service_pb.CleanArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: engine_service_pb.CleanResponse) => void
  ): grpcWeb.ClientReadableStream<engine_service_pb.CleanResponse>;

  getServiceLogs(
    request: engine_service_pb.GetServiceLogsArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<engine_service_pb.GetServiceLogsResponse>;

}

export class EngineServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  getEngineInfo(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<engine_service_pb.GetEngineInfoResponse>;

  createEnclave(
    request: engine_service_pb.CreateEnclaveArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<engine_service_pb.CreateEnclaveResponse>;

  getEnclaves(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<engine_service_pb.GetEnclavesResponse>;

  getExistingAndHistoricalEnclaveIdentifiers(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<engine_service_pb.GetExistingAndHistoricalEnclaveIdentifiersResponse>;

  stopEnclave(
    request: engine_service_pb.StopEnclaveArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  destroyEnclave(
    request: engine_service_pb.DestroyEnclaveArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  clean(
    request: engine_service_pb.CleanArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<engine_service_pb.CleanResponse>;

  getServiceLogs(
    request: engine_service_pb.GetServiceLogsArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<engine_service_pb.GetServiceLogsResponse>;

}

