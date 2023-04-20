// GENERATED CODE -- DO NOT EDIT!

// package: api_container_api
// file: api_container_service.proto

import * as api_container_service_pb from "./api_container_service_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";

interface IApiContainerServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  runStarlarkScript: grpc.MethodDefinition<api_container_service_pb.RunStarlarkScriptArgs, api_container_service_pb.StarlarkRunResponseLine>;
  uploadStarlarkPackage: grpc.MethodDefinition<api_container_service_pb.StreamedDataChunk, google_protobuf_empty_pb.Empty>;
  runStarlarkPackage: grpc.MethodDefinition<api_container_service_pb.RunStarlarkPackageArgs, api_container_service_pb.StarlarkRunResponseLine>;
  startServices: grpc.MethodDefinition<api_container_service_pb.StartServicesArgs, api_container_service_pb.StartServicesResponse>;
  getServices: grpc.MethodDefinition<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  getExistingAndHistoricalServiceIdentifiers: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;
  removeService: grpc.MethodDefinition<api_container_service_pb.RemoveServiceArgs, api_container_service_pb.RemoveServiceResponse>;
  repartition: grpc.MethodDefinition<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.MethodDefinition<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  pauseService: grpc.MethodDefinition<api_container_service_pb.PauseServiceArgs, google_protobuf_empty_pb.Empty>;
  unpauseService: grpc.MethodDefinition<api_container_service_pb.UnpauseServiceArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpGetEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.UploadFilesArtifactArgs, api_container_service_pb.UploadFilesArtifactResponse>;
  uploadFilesArtifactV2: grpc.MethodDefinition<api_container_service_pb.StreamedDataChunk, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.DownloadFilesArtifactResponse>;
  downloadFilesArtifactV2: grpc.MethodDefinition<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact: grpc.MethodDefinition<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.MethodDefinition<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  renderTemplatesToFilesArtifact: grpc.MethodDefinition<api_container_service_pb.RenderTemplatesToFilesArtifactArgs, api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;
  listFilesArtifactNamesAndUuids: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;
}

export const ApiContainerServiceService: IApiContainerServiceService;

export interface IApiContainerServiceServer extends grpc.UntypedServiceImplementation {
  runStarlarkScript: grpc.handleServerStreamingCall<api_container_service_pb.RunStarlarkScriptArgs, api_container_service_pb.StarlarkRunResponseLine>;
  uploadStarlarkPackage: grpc.handleClientStreamingCall<api_container_service_pb.StreamedDataChunk, google_protobuf_empty_pb.Empty>;
  runStarlarkPackage: grpc.handleServerStreamingCall<api_container_service_pb.RunStarlarkPackageArgs, api_container_service_pb.StarlarkRunResponseLine>;
  startServices: grpc.handleUnaryCall<api_container_service_pb.StartServicesArgs, api_container_service_pb.StartServicesResponse>;
  getServices: grpc.handleUnaryCall<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  getExistingAndHistoricalServiceIdentifiers: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;
  removeService: grpc.handleUnaryCall<api_container_service_pb.RemoveServiceArgs, api_container_service_pb.RemoveServiceResponse>;
  repartition: grpc.handleUnaryCall<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.handleUnaryCall<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  pauseService: grpc.handleUnaryCall<api_container_service_pb.PauseServiceArgs, google_protobuf_empty_pb.Empty>;
  unpauseService: grpc.handleUnaryCall<api_container_service_pb.UnpauseServiceArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpGetEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.UploadFilesArtifactArgs, api_container_service_pb.UploadFilesArtifactResponse>;
  uploadFilesArtifactV2: grpc.handleClientStreamingCall<api_container_service_pb.StreamedDataChunk, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.DownloadFilesArtifactResponse>;
  downloadFilesArtifactV2: grpc.handleServerStreamingCall<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.handleUnaryCall<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  renderTemplatesToFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.RenderTemplatesToFilesArtifactArgs, api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;
  listFilesArtifactNamesAndUuids: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;
}

export class ApiContainerServiceClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  runStarlarkScript(argument: api_container_service_pb.RunStarlarkScriptArgs, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;
  runStarlarkScript(argument: api_container_service_pb.RunStarlarkScriptArgs, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;
  uploadStarlarkPackage(callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadStarlarkPackage(metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadStarlarkPackage(metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  runStarlarkPackage(argument: api_container_service_pb.RunStarlarkPackageArgs, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;
  runStarlarkPackage(argument: api_container_service_pb.RunStarlarkPackageArgs, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StarlarkRunResponseLine>;
  startServices(argument: api_container_service_pb.StartServicesArgs, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  startServices(argument: api_container_service_pb.StartServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  startServices(argument: api_container_service_pb.StartServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, callback: grpc.requestCallback<api_container_service_pb.RemoveServiceResponse>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RemoveServiceResponse>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RemoveServiceResponse>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  pauseService(argument: api_container_service_pb.PauseServiceArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  pauseService(argument: api_container_service_pb.PauseServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  pauseService(argument: api_container_service_pb.PauseServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unpauseService(argument: api_container_service_pb.UnpauseServiceArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unpauseService(argument: api_container_service_pb.UnpauseServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unpauseService(argument: api_container_service_pb.UnpauseServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  uploadFilesArtifact(argument: api_container_service_pb.UploadFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientUnaryCall;
  uploadFilesArtifact(argument: api_container_service_pb.UploadFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientUnaryCall;
  uploadFilesArtifact(argument: api_container_service_pb.UploadFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientUnaryCall;
  uploadFilesArtifactV2(callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadFilesArtifactV2(metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadFilesArtifactV2(metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  downloadFilesArtifactV2(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
  downloadFilesArtifactV2(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
}
