// package: api_container_api
// file: api_container_service.proto

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

export class StartExternalContainerRegistrationResponse extends jspb.Message {
  getRegistrationKey(): string;
  setRegistrationKey(value: string): void;

  getIpAddr(): string;
  setIpAddr(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartExternalContainerRegistrationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StartExternalContainerRegistrationResponse): StartExternalContainerRegistrationResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StartExternalContainerRegistrationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartExternalContainerRegistrationResponse;
  static deserializeBinaryFromReader(message: StartExternalContainerRegistrationResponse, reader: jspb.BinaryReader): StartExternalContainerRegistrationResponse;
}

export namespace StartExternalContainerRegistrationResponse {
  export type AsObject = {
    registrationKey: string,
    ipAddr: string,
  }
}

export class FinishExternalContainerRegistrationArgs extends jspb.Message {
  getRegistrationKey(): string;
  setRegistrationKey(value: string): void;

  getContainerId(): string;
  setContainerId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FinishExternalContainerRegistrationArgs.AsObject;
  static toObject(includeInstance: boolean, msg: FinishExternalContainerRegistrationArgs): FinishExternalContainerRegistrationArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FinishExternalContainerRegistrationArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FinishExternalContainerRegistrationArgs;
  static deserializeBinaryFromReader(message: FinishExternalContainerRegistrationArgs, reader: jspb.BinaryReader): FinishExternalContainerRegistrationArgs;
}

export namespace FinishExternalContainerRegistrationArgs {
  export type AsObject = {
    registrationKey: string,
    containerId: string,
  }
}

export class LoadModuleArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): void;

  getContainerImage(): string;
  setContainerImage(value: string): void;

  getSerializedParams(): string;
  setSerializedParams(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: LoadModuleArgs): LoadModuleArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: LoadModuleArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadModuleArgs;
  static deserializeBinaryFromReader(message: LoadModuleArgs, reader: jspb.BinaryReader): LoadModuleArgs;
}

export namespace LoadModuleArgs {
  export type AsObject = {
    moduleId: string,
    containerImage: string,
    serializedParams: string,
  }
}

export class UnloadModuleArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnloadModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: UnloadModuleArgs): UnloadModuleArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: UnloadModuleArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnloadModuleArgs;
  static deserializeBinaryFromReader(message: UnloadModuleArgs, reader: jspb.BinaryReader): UnloadModuleArgs;
}

export namespace UnloadModuleArgs {
  export type AsObject = {
    moduleId: string,
  }
}

export class ExecuteModuleArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): void;

  getSerializedParams(): string;
  setSerializedParams(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteModuleArgs): ExecuteModuleArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecuteModuleArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteModuleArgs;
  static deserializeBinaryFromReader(message: ExecuteModuleArgs, reader: jspb.BinaryReader): ExecuteModuleArgs;
}

export namespace ExecuteModuleArgs {
  export type AsObject = {
    moduleId: string,
    serializedParams: string,
  }
}

export class ExecuteModuleResponse extends jspb.Message {
  getSerializedResult(): string;
  setSerializedResult(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteModuleResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteModuleResponse): ExecuteModuleResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecuteModuleResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteModuleResponse;
  static deserializeBinaryFromReader(message: ExecuteModuleResponse, reader: jspb.BinaryReader): ExecuteModuleResponse;
}

export namespace ExecuteModuleResponse {
  export type AsObject = {
    serializedResult: string,
  }
}

export class GetModuleInfoArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModuleInfoArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetModuleInfoArgs): GetModuleInfoArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetModuleInfoArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetModuleInfoArgs;
  static deserializeBinaryFromReader(message: GetModuleInfoArgs, reader: jspb.BinaryReader): GetModuleInfoArgs;
}

export namespace GetModuleInfoArgs {
  export type AsObject = {
    moduleId: string,
  }
}

export class GetModuleInfoResponse extends jspb.Message {
  getIpAddr(): string;
  setIpAddr(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModuleInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetModuleInfoResponse): GetModuleInfoResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetModuleInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetModuleInfoResponse;
  static deserializeBinaryFromReader(message: GetModuleInfoResponse, reader: jspb.BinaryReader): GetModuleInfoResponse;
}

export namespace GetModuleInfoResponse {
  export type AsObject = {
    ipAddr: string,
  }
}

export class RegisterFilesArtifactsArgs extends jspb.Message {
  getFilesArtifactUrlsMap(): jspb.Map<string, string>;
  clearFilesArtifactUrlsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterFilesArtifactsArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterFilesArtifactsArgs): RegisterFilesArtifactsArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RegisterFilesArtifactsArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterFilesArtifactsArgs;
  static deserializeBinaryFromReader(message: RegisterFilesArtifactsArgs, reader: jspb.BinaryReader): RegisterFilesArtifactsArgs;
}

export namespace RegisterFilesArtifactsArgs {
  export type AsObject = {
    filesArtifactUrlsMap: Array<[string, string]>,
  }
}

export class RegisterServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  getPartitionId(): string;
  setPartitionId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterServiceArgs): RegisterServiceArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RegisterServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterServiceArgs;
  static deserializeBinaryFromReader(message: RegisterServiceArgs, reader: jspb.BinaryReader): RegisterServiceArgs;
}

export namespace RegisterServiceArgs {
  export type AsObject = {
    serviceId: string,
    partitionId: string,
  }
}

export class RegisterServiceResponse extends jspb.Message {
  getIpAddr(): string;
  setIpAddr(value: string): void;

  getRelativeServiceDirpath(): string;
  setRelativeServiceDirpath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterServiceResponse): RegisterServiceResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RegisterServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterServiceResponse;
  static deserializeBinaryFromReader(message: RegisterServiceResponse, reader: jspb.BinaryReader): RegisterServiceResponse;
}

export namespace RegisterServiceResponse {
  export type AsObject = {
    ipAddr: string,
    relativeServiceDirpath: string,
  }
}

export class StartServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  getDockerImage(): string;
  setDockerImage(value: string): void;

  getUsedPortsMap(): jspb.Map<string, boolean>;
  clearUsedPortsMap(): void;
  clearEntrypointArgsList(): void;
  getEntrypointArgsList(): Array<string>;
  setEntrypointArgsList(value: Array<string>): void;
  addEntrypointArgs(value: string, index?: number): string;

  clearCmdArgsList(): void;
  getCmdArgsList(): Array<string>;
  setCmdArgsList(value: Array<string>): void;
  addCmdArgs(value: string, index?: number): string;

  getDockerEnvVarsMap(): jspb.Map<string, string>;
  clearDockerEnvVarsMap(): void;
  getEnclaveDataDirMntDirpath(): string;
  setEnclaveDataDirMntDirpath(value: string): void;

  getFilesArtifactMountDirpathsMap(): jspb.Map<string, string>;
  clearFilesArtifactMountDirpathsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StartServiceArgs): StartServiceArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StartServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServiceArgs;
  static deserializeBinaryFromReader(message: StartServiceArgs, reader: jspb.BinaryReader): StartServiceArgs;
}

export namespace StartServiceArgs {
  export type AsObject = {
    serviceId: string,
    dockerImage: string,
    usedPortsMap: Array<[string, boolean]>,
    entrypointArgsList: Array<string>,
    cmdArgsList: Array<string>,
    dockerEnvVarsMap: Array<[string, string]>,
    enclaveDataDirMntDirpath: string,
    filesArtifactMountDirpathsMap: Array<[string, string]>,
  }
}

export class StartServiceResponse extends jspb.Message {
  getUsedPortsHostPortBindingsMap(): jspb.Map<string, PortBinding>;
  clearUsedPortsHostPortBindingsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StartServiceResponse): StartServiceResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StartServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServiceResponse;
  static deserializeBinaryFromReader(message: StartServiceResponse, reader: jspb.BinaryReader): StartServiceResponse;
}

export namespace StartServiceResponse {
  export type AsObject = {
    usedPortsHostPortBindingsMap: Array<[string, PortBinding.AsObject]>,
  }
}

export class PortBinding extends jspb.Message {
  getInterfaceIp(): string;
  setInterfaceIp(value: string): void;

  getInterfacePort(): string;
  setInterfacePort(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PortBinding.AsObject;
  static toObject(includeInstance: boolean, msg: PortBinding): PortBinding.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PortBinding, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PortBinding;
  static deserializeBinaryFromReader(message: PortBinding, reader: jspb.BinaryReader): PortBinding;
}

export namespace PortBinding {
  export type AsObject = {
    interfaceIp: string,
    interfacePort: string,
  }
}

export class GetServiceInfoArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceInfoArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceInfoArgs): GetServiceInfoArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetServiceInfoArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServiceInfoArgs;
  static deserializeBinaryFromReader(message: GetServiceInfoArgs, reader: jspb.BinaryReader): GetServiceInfoArgs;
}

export namespace GetServiceInfoArgs {
  export type AsObject = {
    serviceId: string,
  }
}

export class GetServiceInfoResponse extends jspb.Message {
  getIpAddr(): string;
  setIpAddr(value: string): void;

  getEnclaveDataDirMountDirpath(): string;
  setEnclaveDataDirMountDirpath(value: string): void;

  getRelativeServiceDirpath(): string;
  setRelativeServiceDirpath(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceInfoResponse): GetServiceInfoResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetServiceInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServiceInfoResponse;
  static deserializeBinaryFromReader(message: GetServiceInfoResponse, reader: jspb.BinaryReader): GetServiceInfoResponse;
}

export namespace GetServiceInfoResponse {
  export type AsObject = {
    ipAddr: string,
    enclaveDataDirMountDirpath: string,
    relativeServiceDirpath: string,
  }
}

export class RemoveServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  getContainerStopTimeoutSeconds(): number;
  setContainerStopTimeoutSeconds(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveServiceArgs): RemoveServiceArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RemoveServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveServiceArgs;
  static deserializeBinaryFromReader(message: RemoveServiceArgs, reader: jspb.BinaryReader): RemoveServiceArgs;
}

export namespace RemoveServiceArgs {
  export type AsObject = {
    serviceId: string,
    containerStopTimeoutSeconds: number,
  }
}

export class RepartitionArgs extends jspb.Message {
  getPartitionServicesMap(): jspb.Map<string, PartitionServices>;
  clearPartitionServicesMap(): void;
  getPartitionConnectionsMap(): jspb.Map<string, PartitionConnections>;
  clearPartitionConnectionsMap(): void;
  hasDefaultConnection(): boolean;
  clearDefaultConnection(): void;
  getDefaultConnection(): PartitionConnectionInfo | undefined;
  setDefaultConnection(value?: PartitionConnectionInfo): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RepartitionArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RepartitionArgs): RepartitionArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RepartitionArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RepartitionArgs;
  static deserializeBinaryFromReader(message: RepartitionArgs, reader: jspb.BinaryReader): RepartitionArgs;
}

export namespace RepartitionArgs {
  export type AsObject = {
    partitionServicesMap: Array<[string, PartitionServices.AsObject]>,
    partitionConnectionsMap: Array<[string, PartitionConnections.AsObject]>,
    defaultConnection?: PartitionConnectionInfo.AsObject,
  }
}

export class PartitionServices extends jspb.Message {
  getServiceIdSetMap(): jspb.Map<string, boolean>;
  clearServiceIdSetMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionServices.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionServices): PartitionServices.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PartitionServices, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PartitionServices;
  static deserializeBinaryFromReader(message: PartitionServices, reader: jspb.BinaryReader): PartitionServices;
}

export namespace PartitionServices {
  export type AsObject = {
    serviceIdSetMap: Array<[string, boolean]>,
  }
}

export class PartitionConnections extends jspb.Message {
  getConnectionInfoMap(): jspb.Map<string, PartitionConnectionInfo>;
  clearConnectionInfoMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionConnections.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionConnections): PartitionConnections.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PartitionConnections, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PartitionConnections;
  static deserializeBinaryFromReader(message: PartitionConnections, reader: jspb.BinaryReader): PartitionConnections;
}

export namespace PartitionConnections {
  export type AsObject = {
    connectionInfoMap: Array<[string, PartitionConnectionInfo.AsObject]>,
  }
}

export class PartitionConnectionInfo extends jspb.Message {
  getIsBlocked(): boolean;
  setIsBlocked(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionConnectionInfo.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionConnectionInfo): PartitionConnectionInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PartitionConnectionInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PartitionConnectionInfo;
  static deserializeBinaryFromReader(message: PartitionConnectionInfo, reader: jspb.BinaryReader): PartitionConnectionInfo;
}

export namespace PartitionConnectionInfo {
  export type AsObject = {
    isBlocked: boolean,
  }
}

export class ExecCommandArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  clearCommandArgsList(): void;
  getCommandArgsList(): Array<string>;
  setCommandArgsList(value: Array<string>): void;
  addCommandArgs(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecCommandArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecCommandArgs): ExecCommandArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecCommandArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecCommandArgs;
  static deserializeBinaryFromReader(message: ExecCommandArgs, reader: jspb.BinaryReader): ExecCommandArgs;
}

export namespace ExecCommandArgs {
  export type AsObject = {
    serviceId: string,
    commandArgsList: Array<string>,
  }
}

export class ExecCommandResponse extends jspb.Message {
  getExitCode(): number;
  setExitCode(value: number): void;

  getLogOutput(): string;
  setLogOutput(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecCommandResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecCommandResponse): ExecCommandResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecCommandResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecCommandResponse;
  static deserializeBinaryFromReader(message: ExecCommandResponse, reader: jspb.BinaryReader): ExecCommandResponse;
}

export namespace ExecCommandResponse {
  export type AsObject = {
    exitCode: number,
    logOutput: string,
  }
}

export class WaitForHttpGetEndpointAvailabilityArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  getPort(): number;
  setPort(value: number): void;

  getPath(): string;
  setPath(value: string): void;

  getInitialDelayMilliseconds(): number;
  setInitialDelayMilliseconds(value: number): void;

  getRetries(): number;
  setRetries(value: number): void;

  getRetriesDelayMilliseconds(): number;
  setRetriesDelayMilliseconds(value: number): void;

  getBodyText(): string;
  setBodyText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WaitForHttpGetEndpointAvailabilityArgs.AsObject;
  static toObject(includeInstance: boolean, msg: WaitForHttpGetEndpointAvailabilityArgs): WaitForHttpGetEndpointAvailabilityArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WaitForHttpGetEndpointAvailabilityArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WaitForHttpGetEndpointAvailabilityArgs;
  static deserializeBinaryFromReader(message: WaitForHttpGetEndpointAvailabilityArgs, reader: jspb.BinaryReader): WaitForHttpGetEndpointAvailabilityArgs;
}

export namespace WaitForHttpGetEndpointAvailabilityArgs {
  export type AsObject = {
    serviceId: string,
    port: number,
    path: string,
    initialDelayMilliseconds: number,
    retries: number,
    retriesDelayMilliseconds: number,
    bodyText: string,
  }
}

export class WaitForHttpPostEndpointAvailabilityArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): void;

  getPort(): number;
  setPort(value: number): void;

  getPath(): string;
  setPath(value: string): void;

  getRequestBody(): string;
  setRequestBody(value: string): void;

  getInitialDelayMilliseconds(): number;
  setInitialDelayMilliseconds(value: number): void;

  getRetries(): number;
  setRetries(value: number): void;

  getRetriesDelayMilliseconds(): number;
  setRetriesDelayMilliseconds(value: number): void;

  getBodyText(): string;
  setBodyText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WaitForHttpPostEndpointAvailabilityArgs.AsObject;
  static toObject(includeInstance: boolean, msg: WaitForHttpPostEndpointAvailabilityArgs): WaitForHttpPostEndpointAvailabilityArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WaitForHttpPostEndpointAvailabilityArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WaitForHttpPostEndpointAvailabilityArgs;
  static deserializeBinaryFromReader(message: WaitForHttpPostEndpointAvailabilityArgs, reader: jspb.BinaryReader): WaitForHttpPostEndpointAvailabilityArgs;
}

export namespace WaitForHttpPostEndpointAvailabilityArgs {
  export type AsObject = {
    serviceId: string,
    port: number,
    path: string,
    requestBody: string,
    initialDelayMilliseconds: number,
    retries: number,
    retriesDelayMilliseconds: number,
    bodyText: string,
  }
}

export class ExecuteBulkCommandsArgs extends jspb.Message {
  getSerializedCommands(): string;
  setSerializedCommands(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteBulkCommandsArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteBulkCommandsArgs): ExecuteBulkCommandsArgs.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ExecuteBulkCommandsArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ExecuteBulkCommandsArgs;
  static deserializeBinaryFromReader(message: ExecuteBulkCommandsArgs, reader: jspb.BinaryReader): ExecuteBulkCommandsArgs;
}

export namespace ExecuteBulkCommandsArgs {
  export type AsObject = {
    serializedCommands: string,
  }
}

export class GetServicesResponse extends jspb.Message {
  getServiceIdsMap(): jspb.Map<string, boolean>;
  clearServiceIdsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServicesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServicesResponse): GetServicesResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetServicesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServicesResponse;
  static deserializeBinaryFromReader(message: GetServicesResponse, reader: jspb.BinaryReader): GetServicesResponse;
}

export namespace GetServicesResponse {
  export type AsObject = {
    serviceIdsMap: Array<[string, boolean]>,
  }
}

export class GetModulesResponse extends jspb.Message {
  getModuleIdsMap(): jspb.Map<string, boolean>;
  clearModuleIdsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModulesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetModulesResponse): GetModulesResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetModulesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetModulesResponse;
  static deserializeBinaryFromReader(message: GetModulesResponse, reader: jspb.BinaryReader): GetModulesResponse;
}

export namespace GetModulesResponse {
  export type AsObject = {
    moduleIdsMap: Array<[string, boolean]>,
  }
}

