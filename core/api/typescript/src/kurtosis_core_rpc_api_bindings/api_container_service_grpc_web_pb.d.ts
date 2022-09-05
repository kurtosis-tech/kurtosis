import * as grpcWeb from 'grpc-web';

import * as api_container_service_pb from './api_container_service_pb';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class ApiContainerServiceClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  loadModule(
    request: api_container_service_pb.LoadModuleArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.LoadModuleResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.LoadModuleResponse>;

  getModules(
    request: api_container_service_pb.GetModulesArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetModulesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetModulesResponse>;

  unloadModule(
    request: api_container_service_pb.UnloadModuleArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.UnloadModuleResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.UnloadModuleResponse>;

  executeModule(
    request: api_container_service_pb.ExecuteModuleArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ExecuteModuleResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ExecuteModuleResponse>;

  registerServices(
    request: api_container_service_pb.RegisterServicesArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.RegisterServicesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.RegisterServicesResponse>;

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

}

export class ApiContainerServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  loadModule(
    request: api_container_service_pb.LoadModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.LoadModuleResponse>;

  getModules(
    request: api_container_service_pb.GetModulesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetModulesResponse>;

  unloadModule(
    request: api_container_service_pb.UnloadModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.UnloadModuleResponse>;

  executeModule(
    request: api_container_service_pb.ExecuteModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ExecuteModuleResponse>;

  registerServices(
    request: api_container_service_pb.RegisterServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.RegisterServicesResponse>;

  startServices(
    request: api_container_service_pb.StartServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.StartServicesResponse>;

  getServices(
    request: api_container_service_pb.GetServicesArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetServicesResponse>;

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

}

