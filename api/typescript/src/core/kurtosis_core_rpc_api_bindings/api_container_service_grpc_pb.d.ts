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
  getServices: grpc.MethodDefinition<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  getExistingAndHistoricalServiceIdentifiers: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;
  execCommand: grpc.MethodDefinition<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  waitForHttpGetEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.MethodDefinition<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.StreamedDataChunk, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.MethodDefinition<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact: grpc.MethodDefinition<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.MethodDefinition<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  listFilesArtifactNamesAndUuids: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;
  inspectFilesArtifactContents: grpc.MethodDefinition<api_container_service_pb.InspectFilesArtifactContentsRequest, api_container_service_pb.InspectFilesArtifactContentsResponse>;
  connectServices: grpc.MethodDefinition<api_container_service_pb.ConnectServicesArgs, api_container_service_pb.ConnectServicesResponse>;
  getStarlarkRun: grpc.MethodDefinition<google_protobuf_empty_pb.Empty, api_container_service_pb.GetStarlarkRunResponse>;
  getStarlarkScriptPlanYaml: grpc.MethodDefinition<api_container_service_pb.StarlarkScriptPlanYamlArgs, api_container_service_pb.PlanYaml>;
  getStarlarkPackagePlanYaml: grpc.MethodDefinition<api_container_service_pb.StarlarkPackagePlanYamlArgs, api_container_service_pb.PlanYaml>;
  createSnapshot: grpc.MethodDefinition<api_container_service_pb.CreateSnapshotArgs, api_container_service_pb.StreamedDataChunk>;
}

export const ApiContainerServiceService: IApiContainerServiceService;

export interface IApiContainerServiceServer extends grpc.UntypedServiceImplementation {
  runStarlarkScript: grpc.handleServerStreamingCall<api_container_service_pb.RunStarlarkScriptArgs, api_container_service_pb.StarlarkRunResponseLine>;
  uploadStarlarkPackage: grpc.handleClientStreamingCall<api_container_service_pb.StreamedDataChunk, google_protobuf_empty_pb.Empty>;
  runStarlarkPackage: grpc.handleServerStreamingCall<api_container_service_pb.RunStarlarkPackageArgs, api_container_service_pb.StarlarkRunResponseLine>;
  getServices: grpc.handleUnaryCall<api_container_service_pb.GetServicesArgs, api_container_service_pb.GetServicesResponse>;
  getExistingAndHistoricalServiceIdentifiers: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>;
  execCommand: grpc.handleUnaryCall<api_container_service_pb.ExecCommandArgs, api_container_service_pb.ExecCommandResponse>;
  waitForHttpGetEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  waitForHttpPostEndpointAvailability: grpc.handleUnaryCall<api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, google_protobuf_empty_pb.Empty>;
  uploadFilesArtifact: grpc.handleClientStreamingCall<api_container_service_pb.StreamedDataChunk, api_container_service_pb.UploadFilesArtifactResponse>;
  downloadFilesArtifact: grpc.handleServerStreamingCall<api_container_service_pb.DownloadFilesArtifactArgs, api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact: grpc.handleUnaryCall<api_container_service_pb.StoreWebFilesArtifactArgs, api_container_service_pb.StoreWebFilesArtifactResponse>;
  storeFilesArtifactFromService: grpc.handleUnaryCall<api_container_service_pb.StoreFilesArtifactFromServiceArgs, api_container_service_pb.StoreFilesArtifactFromServiceResponse>;
  listFilesArtifactNamesAndUuids: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>;
  inspectFilesArtifactContents: grpc.handleUnaryCall<api_container_service_pb.InspectFilesArtifactContentsRequest, api_container_service_pb.InspectFilesArtifactContentsResponse>;
  connectServices: grpc.handleUnaryCall<api_container_service_pb.ConnectServicesArgs, api_container_service_pb.ConnectServicesResponse>;
  getStarlarkRun: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, api_container_service_pb.GetStarlarkRunResponse>;
  getStarlarkScriptPlanYaml: grpc.handleUnaryCall<api_container_service_pb.StarlarkScriptPlanYamlArgs, api_container_service_pb.PlanYaml>;
  getStarlarkPackagePlanYaml: grpc.handleUnaryCall<api_container_service_pb.StarlarkPackagePlanYamlArgs, api_container_service_pb.PlanYaml>;
  createSnapshot: grpc.handleServerStreamingCall<api_container_service_pb.CreateSnapshotArgs, api_container_service_pb.StreamedDataChunk>;
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
  getServices(argument: api_container_service_pb.GetServicesArgs, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getServices(argument: api_container_service_pb.GetServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetServicesResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  getExistingAndHistoricalServiceIdentifiers(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetExistingAndHistoricalServiceIdentifiersResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  execCommand(argument: api_container_service_pb.ExecCommandArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ExecCommandResponse>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpGetEndpointAvailability(argument: api_container_service_pb.WaitForHttpGetEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  waitForHttpPostEndpointAvailability(argument: api_container_service_pb.WaitForHttpPostEndpointAvailabilityArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<google_protobuf_empty_pb.Empty>): grpc.ClientUnaryCall;
  uploadFilesArtifact(callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadFilesArtifact(metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  uploadFilesArtifact(metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.UploadFilesArtifactResponse>): grpc.ClientWritableStream<api_container_service_pb.StreamedDataChunk>;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
  downloadFilesArtifact(argument: api_container_service_pb.DownloadFilesArtifactArgs, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeWebFilesArtifact(argument: api_container_service_pb.StoreWebFilesArtifactArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreWebFilesArtifactResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  storeFilesArtifactFromService(argument: api_container_service_pb.StoreFilesArtifactFromServiceArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.StoreFilesArtifactFromServiceResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
  listFilesArtifactNamesAndUuids(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ListFilesArtifactNamesAndUuidsResponse>): grpc.ClientUnaryCall;
  inspectFilesArtifactContents(argument: api_container_service_pb.InspectFilesArtifactContentsRequest, callback: grpc.requestCallback<api_container_service_pb.InspectFilesArtifactContentsResponse>): grpc.ClientUnaryCall;
  inspectFilesArtifactContents(argument: api_container_service_pb.InspectFilesArtifactContentsRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.InspectFilesArtifactContentsResponse>): grpc.ClientUnaryCall;
  inspectFilesArtifactContents(argument: api_container_service_pb.InspectFilesArtifactContentsRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.InspectFilesArtifactContentsResponse>): grpc.ClientUnaryCall;
  connectServices(argument: api_container_service_pb.ConnectServicesArgs, callback: grpc.requestCallback<api_container_service_pb.ConnectServicesResponse>): grpc.ClientUnaryCall;
  connectServices(argument: api_container_service_pb.ConnectServicesArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ConnectServicesResponse>): grpc.ClientUnaryCall;
  connectServices(argument: api_container_service_pb.ConnectServicesArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.ConnectServicesResponse>): grpc.ClientUnaryCall;
  getStarlarkRun(argument: google_protobuf_empty_pb.Empty, callback: grpc.requestCallback<api_container_service_pb.GetStarlarkRunResponse>): grpc.ClientUnaryCall;
  getStarlarkRun(argument: google_protobuf_empty_pb.Empty, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetStarlarkRunResponse>): grpc.ClientUnaryCall;
  getStarlarkRun(argument: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.GetStarlarkRunResponse>): grpc.ClientUnaryCall;
  getStarlarkScriptPlanYaml(argument: api_container_service_pb.StarlarkScriptPlanYamlArgs, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  getStarlarkScriptPlanYaml(argument: api_container_service_pb.StarlarkScriptPlanYamlArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  getStarlarkScriptPlanYaml(argument: api_container_service_pb.StarlarkScriptPlanYamlArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  getStarlarkPackagePlanYaml(argument: api_container_service_pb.StarlarkPackagePlanYamlArgs, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  getStarlarkPackagePlanYaml(argument: api_container_service_pb.StarlarkPackagePlanYamlArgs, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  getStarlarkPackagePlanYaml(argument: api_container_service_pb.StarlarkPackagePlanYamlArgs, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<api_container_service_pb.PlanYaml>): grpc.ClientUnaryCall;
  createSnapshot(argument: api_container_service_pb.CreateSnapshotArgs, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
  createSnapshot(argument: api_container_service_pb.CreateSnapshotArgs, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<api_container_service_pb.StreamedDataChunk>;
}
