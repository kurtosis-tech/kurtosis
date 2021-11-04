// GENERATED CODE -- DO NOT EDIT!

// package: api_container_api
// file: api_container_service.proto

import * as api_container_service_pb from "./api_container_service_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "grpc";

interface IApiContainerServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  startExternalContainerRegistration: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.StartExternalContainerRegistrationResponse>;
  finishExternalContainerRegistration: grpc.MethodDefinition<api_container_service_pb.FinishExternalContainerRegistrationArgs, google_protobuf_empty_pb.Empty>;
  loadModule: grpc.MethodDefinition<api_container_service_pb.LoadModuleArgs, google_protobuf_empty_pb.Empty>;
  unloadModule: grpc.MethodDefinition<api_container_service_pb.UnloadModuleArgs, google_protobuf_empty_pb.Empty>;
  executeModule: grpc.MethodDefinition<api_container_service_pb.ExecuteModuleArgs, api_container_service_pb.ExecuteModuleResponse>;
  getModuleInfo: grpc.MethodDefinition<api_container_service_pb.GetModuleInfoArgs, api_container_service_pb.GetModuleInfoResponse>;
  registerFilesArtifacts: grpc.MethodDefinition<api_container_service_pb.RegisterFilesArtifactsArgs, google_protobuf_empty_pb.Empty>;
  registerService: grpc.MethodDefinition<api_container_service_pb.RegisterServiceArgs, api_container_service_pb.RegisterServiceResponse>;
  startService: grpc.MethodDefinition<api_container_service_pb.StartServiceArgs, api_container_service_pb.StartServiceResponse>;
  getServiceInfo: grpc.MethodDefinition<api_container_service_pb.GetServiceInfoArgs, api_container_service_pb.GetServiceInfoResponse>;
  removeService: grpc.MethodDefinition<api_container_service_pb.RemoveServiceArgs, google_protobuf_empty_pb.Empty>;
  repartition: grpc.MethodDefinition<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.MethodDefinition<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  waitForHttpGetEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  executeBulkCommands: grpc.MethodDefinition<api_container_service_pb.ExecuteBulkCommandsArgs, google_protobuf_empty_pb.Empty>;
  getServices: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.GetServicesResponse>;
  getModules: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.GetModulesResponse>;
}

export const ApiContainerServiceService: IApiContainerServiceService;

export interface IApiContainerServiceServer extends grpc.UntypedServiceImplementation {
  startExternalContainerRegistration: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.StartExternalContainerRegistrationResponse>;
  finishExternalContainerRegistration: grpc.handleUnaryCall<api_container_service_pb.FinishExternalContainerRegistrationArgs, google_protobuf_empty_pb.Empty>;
  loadModule: grpc.handleUnaryCall<api_container_service_pb.LoadModuleArgs, google_protobuf_empty_pb.Empty>;
  unloadModule: grpc.handleUnaryCall<api_container_service_pb.UnloadModuleArgs, google_protobuf_empty_pb.Empty>;
  executeModule: grpc.handleUnaryCall<api_container_service_pb.ExecuteModuleArgs, api_container_service_pb.ExecuteModuleResponse>;
  getModuleInfo: grpc.handleUnaryCall<api_container_service_pb.GetModuleInfoArgs, api_container_service_pb.GetModuleInfoResponse>;
  registerFilesArtifacts: grpc.handleUnaryCall<api_container_service_pb.RegisterFilesArtifactsArgs, google_protobuf_empty_pb.Empty>;
  registerService: grpc.handleUnaryCall<api_container_service_pb.RegisterServiceArgs, api_container_service_pb.RegisterServiceResponse>;
  startService: grpc.handleUnaryCall<api_container_service_pb.StartServiceArgs, api_container_service_pb.StartServiceResponse>;
  getServiceInfo: grpc.handleUnaryCall<api_container_service_pb.GetServiceInfoArgs, api_container_service_pb.GetServiceInfoResponse>;
  removeService: grpc.handleUnaryCall<api_container_service_pb.RemoveServiceArgs, google_protobuf_empty_pb.Empty>;
  repartition: grpc.handleUnaryCall<api_container_service_pb.RepartitionArgs, google_protobuf_empty_pb.Empty>;
  execCommand: grpc.handleUnaryCall<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  waitForHttpGetEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  executeBulkCommands: grpc.handleUnaryCall<api_container_service_pb.ExecuteBulkCommandsArgs, google_protobuf_empty_pb.Empty>;
  getServices: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.GetServicesResponse>;
  getModules: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.GetModulesResponse>;
}

export class ApiContainerServiceClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  startExternalContainerRegistration(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.StartExternalContainerRegistrationResponse>): grpc.ClientUnaryCall;
  startExternalContainerRegistration(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartExternalContainerRegistrationResponse>): grpc.ClientUnaryCall;
  startExternalContainerRegistration(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartExternalContainerRegistrationResponse>): grpc.ClientUnaryCall;
  finishExternalContainerRegistration(argument: api_container_service_pb.FinishExternalContainerRegistrationArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  finishExternalContainerRegistration(argument: api_container_service_pb.FinishExternalContainerRegistrationArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  finishExternalContainerRegistration(argument: api_container_service_pb.FinishExternalContainerRegistrationArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  loadModule(argument: api_container_service_pb.LoadModuleArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  loadModule(argument: api_container_service_pb.LoadModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  loadModule(argument: api_container_service_pb.LoadModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  unloadModule(argument: api_container_service_pb.UnloadModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  executeModule(argument: api_container_service_pb.ExecuteModuleArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecuteModuleResponse>): grpc.ClientUnaryCall;
  getModuleInfo(argument: api_container_service_pb.GetModuleInfoArgs, callback: grpc.requestCallback<api_container_service_pb.GetModuleInfoResponse>): grpc.ClientUnaryCall;
  getModuleInfo(argument: api_container_service_pb.GetModuleInfoArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModuleInfoResponse>): grpc.ClientUnaryCall;
  getModuleInfo(argument: api_container_service_pb.GetModuleInfoArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModuleInfoResponse>): grpc.ClientUnaryCall;
  registerFilesArtifacts(argument: api_container_service_pb.RegisterFilesArtifactsArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  registerFilesArtifacts(argument: api_container_service_pb.RegisterFilesArtifactsArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  registerFilesArtifacts(argument: api_container_service_pb.RegisterFilesArtifactsArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  registerService(argument: api_container_service_pb.RegisterServiceArgs, callback: grpc.requestCallback<api_container_service_pb.RegisterServiceResponse>): grpc.ClientUnaryCall;
  registerService(argument: api_container_service_pb.RegisterServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RegisterServiceResponse>): grpc.ClientUnaryCall;
  registerService(argument: api_container_service_pb.RegisterServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.RegisterServiceResponse>): grpc.ClientUnaryCall;
  startService(argument: api_container_service_pb.StartServiceArgs, callback: grpc.requestCallback<api_container_service_pb.StartServiceResponse>): grpc.ClientUnaryCall;
  startService(argument: api_container_service_pb.StartServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServiceResponse>): grpc.ClientUnaryCall;
  startService(argument: api_container_service_pb.StartServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StartServiceResponse>): grpc.ClientUnaryCall;
  getServiceInfo(argument: api_container_service_pb.GetServiceInfoArgs, callback: grpc.requestCallback<api_container_service_pb.GetServiceInfoResponse>): grpc.ClientUnaryCall;
  getServiceInfo(argument: api_container_service_pb.GetServiceInfoArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServiceInfoResponse>): grpc.ClientUnaryCall;
  getServiceInfo(argument: api_container_service_pb.GetServiceInfoArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServiceInfoResponse>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  removeService(argument: api_container_service_pb.RemoveServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  repartition(argument: api_container_service_pb.RepartitionArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  executeBulkCommands(argument: api_container_service_pb.ExecuteBulkCommandsArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  executeBulkCommands(argument: api_container_service_pb.ExecuteBulkCommandsArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  executeBulkCommands(argument: api_container_service_pb.ExecuteBulkCommandsArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  getServices(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getModules(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
  getModules(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
  getModules(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetModulesResponse>): grpc.ClientUnaryCall;
}
