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

  unloadModule(
    request: api_container_service_pb.UnloadModuleArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  executeModule(
    request: api_container_service_pb.ExecuteModuleArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.ExecuteModuleResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.ExecuteModuleResponse>;

  getModuleInfo(
    request: api_container_service_pb.GetModuleInfoArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetModuleInfoResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetModuleInfoResponse>;

  registerFilesArtifacts(
    request: api_container_service_pb.RegisterFilesArtifactsArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

  registerService(
    request: api_container_service_pb.RegisterServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.RegisterServiceResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.RegisterServiceResponse>;

  startService(
    request: api_container_service_pb.StartServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.StartServiceResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.StartServiceResponse>;

  getServiceInfo(
    request: api_container_service_pb.GetServiceInfoArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetServiceInfoResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetServiceInfoResponse>;

  removeService(
    request: api_container_service_pb.RemoveServiceArgs,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: google_protobuf_empty_pb.Empty) => void
  ): grpcWeb.ClientReadableStream<google_protobuf_empty_pb.Empty>;

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

  getServices(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetServicesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetServicesResponse>;

  getModules(
    request: google_protobuf_empty_pb.Empty,
    metadata: grpcWeb.Metadata | undefined,
    callback: (err: grpcWeb.RpcError,
               response: api_container_service_pb.GetModulesResponse) => void
  ): grpcWeb.ClientReadableStream<api_container_service_pb.GetModulesResponse>;

}

export class ApiContainerServicePromiseClient {
  constructor (hostname: string,
               credentials?: null | { [index: string]: string; },
               options?: null | { [index: string]: any; });

  loadModule(
    request: api_container_service_pb.LoadModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.LoadModuleResponse>;

  unloadModule(
    request: api_container_service_pb.UnloadModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  executeModule(
    request: api_container_service_pb.ExecuteModuleArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.ExecuteModuleResponse>;

  getModuleInfo(
    request: api_container_service_pb.GetModuleInfoArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetModuleInfoResponse>;

  registerFilesArtifacts(
    request: api_container_service_pb.RegisterFilesArtifactsArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

  registerService(
    request: api_container_service_pb.RegisterServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.RegisterServiceResponse>;

  startService(
    request: api_container_service_pb.StartServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.StartServiceResponse>;

  getServiceInfo(
    request: api_container_service_pb.GetServiceInfoArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetServiceInfoResponse>;

  removeService(
    request: api_container_service_pb.RemoveServiceArgs,
    metadata?: grpcWeb.Metadata
  ): Promise<google_protobuf_empty_pb.Empty>;

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

  getServices(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetServicesResponse>;

  getModules(
    request: google_protobuf_empty_pb.Empty,
    metadata?: grpcWeb.Metadata
  ): Promise<api_container_service_pb.GetModulesResponse>;

}

