import * as grpcWeb from 'grpc-web';

import * as api_container_service_pb from './api_container_service_pb';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


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

  startServices(
    request: api_container_service_pb.StartServicesArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.StartServicesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StartServicesResponse>;

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

  removeService(
    request: api_container_service_pb.RemoveServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.RemoveServiceResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.RemoveServiceResponse>;

  repartition(
    request: api_container_service_pb.RepartitionArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  execCommand(
    request: api_container_service_pb.ExecCommandArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ExecCommandResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ExecCommandResponse>;

  pauseService(
    request: api_container_service_pb.PauseServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  unpauseService(
    request: api_container_service_pb.UnpauseServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

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

  uploadFilesArtifact(
    request: api_container_service_pb.UploadFilesArtifactArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.UploadFilesArtifactResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.UploadFilesArtifactResponse>;

  downloadFilesArtifact(
    request: api_container_service_pb.DownloadFilesArtifactArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.DownloadFilesArtifactResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.DownloadFilesArtifactResponse>;

  downloadFilesArtifactV2(
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

  renderTemplatesToFilesArtifact(
    request: api_container_service_pb.RenderTemplatesToFilesArtifactArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.RenderTemplatesToFilesArtifactResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;

  listFilesArtifactNamesAndUuids(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;

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

  startServices(
    request: api_container_service_pb.StartServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.StartServicesResponse>;

  getServices(
    request: api_container_service_pb.GetServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetServicesResponse>;

  getExistingAndHistoricalServiceIdentifiers(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;

  removeService(
    request: api_container_service_pb.RemoveServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.RemoveServiceResponse>;

  repartition(
    request: api_container_service_pb.RepartitionArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  execCommand(
    request: api_container_service_pb.ExecCommandArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ExecCommandResponse>;

  pauseService(
    request: api_container_service_pb.PauseServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  unpauseService(
    request: api_container_service_pb.UnpauseServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  waitForHttpGetEndpointAvailability(
    request: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  waitForHttpPostEndpointAvailability(
    request: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  uploadFilesArtifact(
    request: api_container_service_pb.UploadFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.UploadFilesArtifactResponse>;

  downloadFilesArtifact(
    request: api_container_service_pb.DownloadFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.DownloadFilesArtifactResponse>;

  downloadFilesArtifactV2(
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

  renderTemplatesToFilesArtifact(
    request: api_container_service_pb.RenderTemplatesToFilesArtifactArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;

  listFilesArtifactNamesAndUuids(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;

}

