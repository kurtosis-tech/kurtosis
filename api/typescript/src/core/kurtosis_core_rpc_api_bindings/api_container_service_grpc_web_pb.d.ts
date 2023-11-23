import * as grpcWeb from 'grpc-web';

import * as api_container_service_pb from './api_container_service_pb'; // proto import: "api_container_service.proto"
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb'; // proto import: "google/protobuf/empty.proto"


export class ApiContainerServiceClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  runStarlarkScript(
    request: api_container_service_pb.RunStarlarkScriptArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;

  runStarlarkPackage(
    request: api_container_service_pb.RunStarlarkPackageArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;

  getServices(
    request: api_container_service_pb.GetServicesArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetServicesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetServicesResponse>;

  getExistingAndHistoricalServiceIdentifiers(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;

  execCommand(
    request: api_container_service_pb.ExecCommandArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ExecCommandResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ExecCommandResponse>;

  waitForHttpGetEndpointAvailability(
    request: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  waitForHttpPostEndpointAvailability(
    request: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  downloadFilesArtifact(
    request: api_container_service_pb.DownloadFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;

  storeWebFilesArtifact(
    request: api_container_service_pb.StoreWebFilesArtifactArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.StoreWebFilesArtifactResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StoreWebFilesArtifactResponse>;

  storeFilesArtifactFromService(
    request: api_container_service_pb.StoreFilesArtifactFromServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.StoreFilesArtifactFromServiceResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StoreFilesArtifactFromServiceResponse>;

  listFilesArtifactNamesAndUuids(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;

  inspectFilesArtifactContents(
    request: api_container_service_pb.InspectFilesArtifactContentsRequest,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.InspectFilesArtifactContentsResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.InspectFilesArtifactContentsResponse>;

  connectServices(
    request: api_container_service_pb.ConnectServicesArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ConnectServicesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ConnectServicesResponse>;

  getStarlarkRun(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetStarlarkRunResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetStarlarkRunResponse>;

}

export class ApiContainerServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  runStarlarkScript(
    request: api_container_service_pb.RunStarlarkScriptArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;

  runStarlarkPackage(
    request: api_container_service_pb.RunStarlarkPackageArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;

  getServices(
    request: api_container_service_pb.GetServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetServicesResponse>;

  getExistingAndHistoricalServiceIdentifiers(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;

  execCommand(
    request: api_container_service_pb.ExecCommandArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ExecCommandResponse>;

  waitForHttpGetEndpointAvailability(
    request: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  waitForHttpPostEndpointAvailability(
    request: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  downloadFilesArtifact(
    request: api_container_service_pb.DownloadFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;

  storeWebFilesArtifact(
    request: api_container_service_pb.StoreWebFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.StoreWebFilesArtifactResponse>;

  storeFilesArtifactFromService(
    request: api_container_service_pb.StoreFilesArtifactFromServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.StoreFilesArtifactFromServiceResponse>;

  listFilesArtifactNamesAndUuids(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;

  inspectFilesArtifactContents(
    request: api_container_service_pb.InspectFilesArtifactContentsRequest,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.InspectFilesArtifactContentsResponse>;

  connectServices(
    request: api_container_service_pb.ConnectServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ConnectServicesResponse>;

  getStarlarkRun(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetStarlarkRunResponse>;

}

