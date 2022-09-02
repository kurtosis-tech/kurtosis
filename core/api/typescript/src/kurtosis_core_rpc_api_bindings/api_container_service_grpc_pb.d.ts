// GENERATED CODE -- DO NOT EDIT!

// package: api_container_api
// file: api_container_service.proto

import * as api_container_service_pb from "./api_container_service_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";

interface IApiContainerServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  loadModule: grpc.MethodDefinition<api_container_service_pb.LoadModuleArgs, api_container_service_pb.LoadModuleResponse>;
  getModules: grpc.MethodDefinition<api_container_service_pb.GetModulesArgs, api_container_service_pb.GetModulesResponse>;
  unloadModule: grpc.MethodDefinition<api_container_service_pb.UnloadModuleArgs, api_container_service_pb.UnloadModuleResponse>;
  executeModule: grpc.MethodDefinition<api_container_service_pb.ExecuteModuleArgs, api_container_service_pb.ExecuteModuleResponse>;
  registerServices: grpc.MethodDefinition<api_container_service_pb.RegisterServicesArgs, api_container_service_pb.RegisterServicesResponse>;
  startServices: grpc.MethodDefinition<api_container_service_pb.StartServicesArgs, api_container_service_pb.StartServicesResponse>;
  getServices: grpc.MethodDefinition<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  removeService: grpc.MethodDefinition<api_container_service_pb.RemoveServiceArgs, api_container_service_pb.RemoveServiceResponse>;
  repartition: grpc.MethodDefinition<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.MethodDefinition<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  pauseService: grpc.MethodDefinition<api_container_service_pb.PauseServiceArgs, google_protobuf_empty_pb.Empty>;
  unpauseService: grpc.MethodDefinition<api_container_service_pb.UnpauseServiceArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpGetEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.UploadFilesArtifactArgs, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.DownloadFilesArtifactResponse>;
  storeWebFilesArtifact: grpc.MethodDefinition<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.MethodDefinition<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  renderTemplatesToFilesArtifact: grpc.MethodDefinition<api_container_service_pb.RenderTemplatesToFilesArtifactArgs, api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;
}

export const ApiContainerServiceService: IApiContainerServiceService;

export interface IApiContainerServiceServer extends grpc.UntypedServiceImplementation {
  loadModule: grpc.handleUnaryCall<api_container_service_pb.LoadModuleArgs, api_container_service_pb.LoadModuleResponse>;
  getModules: grpc.handleUnaryCall<api_container_service_pb.GetModulesArgs, api_container_service_pb.GetModulesResponse>;
  unloadModule: grpc.handleUnaryCall<api_container_service_pb.UnloadModuleArgs, api_container_service_pb.UnloadModuleResponse>;
  executeModule: grpc.handleUnaryCall<api_container_service_pb.ExecuteModuleArgs, api_container_service_pb.ExecuteModuleResponse>;
  registerServices: grpc.handleUnaryCall<api_container_service_pb.RegisterServicesArgs, api_container_service_pb.RegisterServicesResponse>;
  startServices: grpc.handleUnaryCall<api_container_service_pb.StartServicesArgs, api_container_service_pb.StartServicesResponse>;
  getServices: grpc.handleUnaryCall<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  removeService: grpc.handleUnaryCall<api_container_service_pb.RemoveServiceArgs, api_container_service_pb.RemoveServiceResponse>;
  repartition: grpc.handleUnaryCall<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.handleUnaryCall<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  pauseService: grpc.handleUnaryCall<api_container_service_pb.PauseServiceArgs, google_protobuf_empty_pb.Empty>;
  unpauseService: grpc.handleUnaryCall<api_container_service_pb.UnpauseServiceArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpGetEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.UploadFilesArtifactArgs, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.DownloadFilesArtifactResponse>;
  storeWebFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.handleUnaryCall<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  renderTemplatesToFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.RenderTemplatesToFilesArtifactArgs, api_container_service_pb.RenderTemplatesToFilesArtifactResponse>;
}

export class ApiContainerServiceClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  loadModule(argument: api_container_service_pb.LoadModuleArgs, callback: grpc.requestCallback<api_container_service_pb.LoadModuleResponse>): grpc.ClientUnaryCall;
  loadModule(argument: api_container_service_pb.LoadModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.LoadModuleResponse>): grpc.ClientUnaryCall;
  loadModule(argument: api_container_service_pb.LoadModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.LoadModuleResponse>): grpc.ClientUnaryCall;
  getModules(argument: api_container_service_pb.GetModulesArgs, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
  getModules(argument: api_container_service_pb.GetModulesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
  getModules(argument: api_container_service_pb.GetModulesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, callback: grpc.requestCallback<api_container_service_pb.UnloadModuleResponse>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UnloadModuleResponse>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UnloadModuleResponse>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  registerServices(argument: api_container_service_pb.RegisterServicesArgs, callback: grpc.requestCallback<api_container_service_pb.RegisterServicesResponse>): grpc.ClientUnaryCall;
  registerServices(argument: api_container_service_pb.RegisterServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RegisterServicesResponse>): grpc.ClientUnaryCall;
  registerServices(argument: api_container_service_pb.RegisterServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RegisterServicesResponse>): grpc.ClientUnaryCall;
  startServices(argument: api_container_service_pb.StartServicesArgs, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  startServices(argument: api_container_service_pb.StartServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  startServices(argument: api_container_service_pb.StartServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
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
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.DownloadFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
  renderTemplatesToFilesArtifact(argument: api_container_service_pb.RenderTemplatesToFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RenderTemplatesToFilesArtifactResponse>): grpc.ClientUnaryCall;
}
