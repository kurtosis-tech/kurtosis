import * as jspb from 'google-protobuf'

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class Port extends jspb.Message {
  getNumber(): number;
  setNumber(value: number): Port;

  getProtocol(): Port.Protocol;
  setProtocol(value: Port.Protocol): Port;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Port.AsObject;
  static toObject(includeInstance: boolean, msg: Port): Port.AsObject;
  static serializeBinaryToWriter(message: Port, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Port;
  static deserializeBinaryFromReader(message: Port, reader: jspb.BinaryReader): Port;
}

export namespace Port {
  export type AsObject = {
    number: number,
    protocol: Port.Protocol,
  }

  export enum Protocol { 
    TCP = 0,
    SCTP = 1,
    UDP = 2,
  }
}

export class LoadModuleArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): LoadModuleArgs;

  getContainerImage(): string;
  setContainerImage(value: string): LoadModuleArgs;

  getSerializedParams(): string;
  setSerializedParams(value: string): LoadModuleArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: LoadModuleArgs): LoadModuleArgs.AsObject;
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

export class LoadModuleResponse extends jspb.Message {
  getPrivateIpAddr(): string;
  setPrivateIpAddr(value: string): LoadModuleResponse;

  getPrivatePort(): Port | undefined;
  setPrivatePort(value?: Port): LoadModuleResponse;
  hasPrivatePort(): boolean;
  clearPrivatePort(): LoadModuleResponse;

  getPublicIpAddr(): string;
  setPublicIpAddr(value: string): LoadModuleResponse;

  getPublicPort(): Port | undefined;
  setPublicPort(value?: Port): LoadModuleResponse;
  hasPublicPort(): boolean;
  clearPublicPort(): LoadModuleResponse;

  getGuid(): string;
  setGuid(value: string): LoadModuleResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadModuleResponse.AsObject;
  static toObject(includeInstance: boolean, msg: LoadModuleResponse): LoadModuleResponse.AsObject;
  static serializeBinaryToWriter(message: LoadModuleResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadModuleResponse;
  static deserializeBinaryFromReader(message: LoadModuleResponse, reader: jspb.BinaryReader): LoadModuleResponse;
}

export namespace LoadModuleResponse {
  export type AsObject = {
    privateIpAddr: string,
    privatePort?: Port.AsObject,
    publicIpAddr: string,
    publicPort?: Port.AsObject,
    guid: string,
  }
}

export class UnloadModuleArgs extends jspb.Message {
  getModuleId(): string;
  setModuleId(value: string): UnloadModuleArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnloadModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: UnloadModuleArgs): UnloadModuleArgs.AsObject;
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
  setModuleId(value: string): ExecuteModuleArgs;

  getSerializedParams(): string;
  setSerializedParams(value: string): ExecuteModuleArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteModuleArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteModuleArgs): ExecuteModuleArgs.AsObject;
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
  setSerializedResult(value: string): ExecuteModuleResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecuteModuleResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecuteModuleResponse): ExecuteModuleResponse.AsObject;
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
  setModuleId(value: string): GetModuleInfoArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModuleInfoArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetModuleInfoArgs): GetModuleInfoArgs.AsObject;
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
  getPrivateIpAddr(): string;
  setPrivateIpAddr(value: string): GetModuleInfoResponse;

  getPrivatePort(): Port | undefined;
  setPrivatePort(value?: Port): GetModuleInfoResponse;
  hasPrivatePort(): boolean;
  clearPrivatePort(): GetModuleInfoResponse;

  getPublicIpAddr(): string;
  setPublicIpAddr(value: string): GetModuleInfoResponse;

  getPublicPort(): Port | undefined;
  setPublicPort(value?: Port): GetModuleInfoResponse;
  hasPublicPort(): boolean;
  clearPublicPort(): GetModuleInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModuleInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetModuleInfoResponse): GetModuleInfoResponse.AsObject;
  static serializeBinaryToWriter(message: GetModuleInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetModuleInfoResponse;
  static deserializeBinaryFromReader(message: GetModuleInfoResponse, reader: jspb.BinaryReader): GetModuleInfoResponse;
}

export namespace GetModuleInfoResponse {
  export type AsObject = {
    privateIpAddr: string,
    privatePort?: Port.AsObject,
    publicIpAddr: string,
    publicPort?: Port.AsObject,
  }
}

export class RegisterServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): RegisterServiceArgs;

  getPartitionId(): string;
  setPartitionId(value: string): RegisterServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterServiceArgs): RegisterServiceArgs.AsObject;
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
  getPrivateIpAddr(): string;
  setPrivateIpAddr(value: string): RegisterServiceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterServiceResponse): RegisterServiceResponse.AsObject;
  static serializeBinaryToWriter(message: RegisterServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterServiceResponse;
  static deserializeBinaryFromReader(message: RegisterServiceResponse, reader: jspb.BinaryReader): RegisterServiceResponse;
}

export namespace RegisterServiceResponse {
  export type AsObject = {
    privateIpAddr: string,
  }
}

export class StartServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): StartServiceArgs;

  getDockerImage(): string;
  setDockerImage(value: string): StartServiceArgs;

  getPrivatePortsMap(): jspb.Map<string, Port>;
  clearPrivatePortsMap(): StartServiceArgs;

  getEntrypointArgsList(): Array<string>;
  setEntrypointArgsList(value: Array<string>): StartServiceArgs;
  clearEntrypointArgsList(): StartServiceArgs;
  addEntrypointArgs(value: string, index?: number): StartServiceArgs;

  getCmdArgsList(): Array<string>;
  setCmdArgsList(value: Array<string>): StartServiceArgs;
  clearCmdArgsList(): StartServiceArgs;
  addCmdArgs(value: string, index?: number): StartServiceArgs;

  getDockerEnvVarsMap(): jspb.Map<string, string>;
  clearDockerEnvVarsMap(): StartServiceArgs;

  getFilesArtifactMountpointsMap(): jspb.Map<string, string>;
  clearFilesArtifactMountpointsMap(): StartServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StartServiceArgs): StartServiceArgs.AsObject;
  static serializeBinaryToWriter(message: StartServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServiceArgs;
  static deserializeBinaryFromReader(message: StartServiceArgs, reader: jspb.BinaryReader): StartServiceArgs;
}

export namespace StartServiceArgs {
  export type AsObject = {
    serviceId: string,
    dockerImage: string,
    privatePortsMap: Array<[string, Port.AsObject]>,
    entrypointArgsList: Array<string>,
    cmdArgsList: Array<string>,
    dockerEnvVarsMap: Array<[string, string]>,
    filesArtifactMountpointsMap: Array<[string, string]>,
  }
}

export class StartServiceResponse extends jspb.Message {
  getPublicIpAddr(): string;
  setPublicIpAddr(value: string): StartServiceResponse;

  getPublicPortsMap(): jspb.Map<string, Port>;
  clearPublicPortsMap(): StartServiceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StartServiceResponse): StartServiceResponse.AsObject;
  static serializeBinaryToWriter(message: StartServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServiceResponse;
  static deserializeBinaryFromReader(message: StartServiceResponse, reader: jspb.BinaryReader): StartServiceResponse;
}

export namespace StartServiceResponse {
  export type AsObject = {
    publicIpAddr: string,
    publicPortsMap: Array<[string, Port.AsObject]>,
  }
}

export class GetServiceInfoArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): GetServiceInfoArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceInfoArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceInfoArgs): GetServiceInfoArgs.AsObject;
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
  getPrivateIpAddr(): string;
  setPrivateIpAddr(value: string): GetServiceInfoResponse;

  getPrivatePortsMap(): jspb.Map<string, Port>;
  clearPrivatePortsMap(): GetServiceInfoResponse;

  getPublicIpAddr(): string;
  setPublicIpAddr(value: string): GetServiceInfoResponse;

  getPublicPortsMap(): jspb.Map<string, Port>;
  clearPublicPortsMap(): GetServiceInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceInfoResponse): GetServiceInfoResponse.AsObject;
  static serializeBinaryToWriter(message: GetServiceInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServiceInfoResponse;
  static deserializeBinaryFromReader(message: GetServiceInfoResponse, reader: jspb.BinaryReader): GetServiceInfoResponse;
}

export namespace GetServiceInfoResponse {
  export type AsObject = {
    privateIpAddr: string,
    privatePortsMap: Array<[string, Port.AsObject]>,
    publicIpAddr: string,
    publicPortsMap: Array<[string, Port.AsObject]>,
  }
}

export class RemoveServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): RemoveServiceArgs;

  getContainerStopTimeoutSeconds(): number;
  setContainerStopTimeoutSeconds(value: number): RemoveServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveServiceArgs): RemoveServiceArgs.AsObject;
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
  clearPartitionServicesMap(): RepartitionArgs;

  getPartitionConnectionsMap(): jspb.Map<string, PartitionConnections>;
  clearPartitionConnectionsMap(): RepartitionArgs;

  getDefaultConnection(): PartitionConnectionInfo | undefined;
  setDefaultConnection(value?: PartitionConnectionInfo): RepartitionArgs;
  hasDefaultConnection(): boolean;
  clearDefaultConnection(): RepartitionArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RepartitionArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RepartitionArgs): RepartitionArgs.AsObject;
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
  clearServiceIdSetMap(): PartitionServices;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionServices.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionServices): PartitionServices.AsObject;
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
  clearConnectionInfoMap(): PartitionConnections;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionConnections.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionConnections): PartitionConnections.AsObject;
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
  getPacketLossPercentage(): number;
  setPacketLossPercentage(value: number): PartitionConnectionInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionConnectionInfo.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionConnectionInfo): PartitionConnectionInfo.AsObject;
  static serializeBinaryToWriter(message: PartitionConnectionInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PartitionConnectionInfo;
  static deserializeBinaryFromReader(message: PartitionConnectionInfo, reader: jspb.BinaryReader): PartitionConnectionInfo;
}

export namespace PartitionConnectionInfo {
  export type AsObject = {
    packetLossPercentage: number,
  }
}

export class ExecCommandArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): ExecCommandArgs;

  getCommandArgsList(): Array<string>;
  setCommandArgsList(value: Array<string>): ExecCommandArgs;
  clearCommandArgsList(): ExecCommandArgs;
  addCommandArgs(value: string, index?: number): ExecCommandArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecCommandArgs.AsObject;
  static toObject(includeInstance: boolean, msg: ExecCommandArgs): ExecCommandArgs.AsObject;
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

export class PauseServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): PauseServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: PauseServiceArgs): PauseServiceArgs.AsObject;
  static serializeBinaryToWriter(message: PauseServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseServiceArgs;
  static deserializeBinaryFromReader(message: PauseServiceArgs, reader: jspb.BinaryReader): PauseServiceArgs;
}

export namespace PauseServiceArgs {
  export type AsObject = {
    serviceId: string,
  }
}

export class UnpauseServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): UnpauseServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnpauseServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: UnpauseServiceArgs): UnpauseServiceArgs.AsObject;
  static serializeBinaryToWriter(message: UnpauseServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnpauseServiceArgs;
  static deserializeBinaryFromReader(message: UnpauseServiceArgs, reader: jspb.BinaryReader): UnpauseServiceArgs;
}

export namespace UnpauseServiceArgs {
  export type AsObject = {
    serviceId: string,
  }
}

export class ExecCommandResponse extends jspb.Message {
  getExitCode(): number;
  setExitCode(value: number): ExecCommandResponse;

  getLogOutput(): string;
  setLogOutput(value: string): ExecCommandResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ExecCommandResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ExecCommandResponse): ExecCommandResponse.AsObject;
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
  setServiceId(value: string): WaitForHttpGetEndpointAvailabilityArgs;

  getPort(): number;
  setPort(value: number): WaitForHttpGetEndpointAvailabilityArgs;

  getPath(): string;
  setPath(value: string): WaitForHttpGetEndpointAvailabilityArgs;

  getInitialDelayMilliseconds(): number;
  setInitialDelayMilliseconds(value: number): WaitForHttpGetEndpointAvailabilityArgs;

  getRetries(): number;
  setRetries(value: number): WaitForHttpGetEndpointAvailabilityArgs;

  getRetriesDelayMilliseconds(): number;
  setRetriesDelayMilliseconds(value: number): WaitForHttpGetEndpointAvailabilityArgs;

  getBodyText(): string;
  setBodyText(value: string): WaitForHttpGetEndpointAvailabilityArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WaitForHttpGetEndpointAvailabilityArgs.AsObject;
  static toObject(includeInstance: boolean, msg: WaitForHttpGetEndpointAvailabilityArgs): WaitForHttpGetEndpointAvailabilityArgs.AsObject;
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
  setServiceId(value: string): WaitForHttpPostEndpointAvailabilityArgs;

  getPort(): number;
  setPort(value: number): WaitForHttpPostEndpointAvailabilityArgs;

  getPath(): string;
  setPath(value: string): WaitForHttpPostEndpointAvailabilityArgs;

  getRequestBody(): string;
  setRequestBody(value: string): WaitForHttpPostEndpointAvailabilityArgs;

  getInitialDelayMilliseconds(): number;
  setInitialDelayMilliseconds(value: number): WaitForHttpPostEndpointAvailabilityArgs;

  getRetries(): number;
  setRetries(value: number): WaitForHttpPostEndpointAvailabilityArgs;

  getRetriesDelayMilliseconds(): number;
  setRetriesDelayMilliseconds(value: number): WaitForHttpPostEndpointAvailabilityArgs;

  getBodyText(): string;
  setBodyText(value: string): WaitForHttpPostEndpointAvailabilityArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WaitForHttpPostEndpointAvailabilityArgs.AsObject;
  static toObject(includeInstance: boolean, msg: WaitForHttpPostEndpointAvailabilityArgs): WaitForHttpPostEndpointAvailabilityArgs.AsObject;
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

export class GetServicesResponse extends jspb.Message {
  getServiceIdsMap(): jspb.Map<string, boolean>;
  clearServiceIdsMap(): GetServicesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServicesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServicesResponse): GetServicesResponse.AsObject;
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
  clearModuleIdsMap(): GetModulesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetModulesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetModulesResponse): GetModulesResponse.AsObject;
  static serializeBinaryToWriter(message: GetModulesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetModulesResponse;
  static deserializeBinaryFromReader(message: GetModulesResponse, reader: jspb.BinaryReader): GetModulesResponse;
}

export namespace GetModulesResponse {
  export type AsObject = {
    moduleIdsMap: Array<[string, boolean]>,
  }
}

export class UploadFilesArtifactArgs extends jspb.Message {
  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): UploadFilesArtifactArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UploadFilesArtifactArgs.AsObject;
  static toObject(includeInstance: boolean, msg: UploadFilesArtifactArgs): UploadFilesArtifactArgs.AsObject;
  static serializeBinaryToWriter(message: UploadFilesArtifactArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UploadFilesArtifactArgs;
  static deserializeBinaryFromReader(message: UploadFilesArtifactArgs, reader: jspb.BinaryReader): UploadFilesArtifactArgs;
}

export namespace UploadFilesArtifactArgs {
  export type AsObject = {
    data: Uint8Array | string,
  }
}

export class UploadFilesArtifactResponse extends jspb.Message {
  getUuid(): string;
  setUuid(value: string): UploadFilesArtifactResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UploadFilesArtifactResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UploadFilesArtifactResponse): UploadFilesArtifactResponse.AsObject;
  static serializeBinaryToWriter(message: UploadFilesArtifactResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UploadFilesArtifactResponse;
  static deserializeBinaryFromReader(message: UploadFilesArtifactResponse, reader: jspb.BinaryReader): UploadFilesArtifactResponse;
}

export namespace UploadFilesArtifactResponse {
  export type AsObject = {
    uuid: string,
  }
}

export class StoreWebFilesArtifactArgs extends jspb.Message {
  getUrl(): string;
  setUrl(value: string): StoreWebFilesArtifactArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreWebFilesArtifactArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StoreWebFilesArtifactArgs): StoreWebFilesArtifactArgs.AsObject;
  static serializeBinaryToWriter(message: StoreWebFilesArtifactArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreWebFilesArtifactArgs;
  static deserializeBinaryFromReader(message: StoreWebFilesArtifactArgs, reader: jspb.BinaryReader): StoreWebFilesArtifactArgs;
}

export namespace StoreWebFilesArtifactArgs {
  export type AsObject = {
    url: string,
  }
}

export class StoreWebFilesArtifactResponse extends jspb.Message {
  getUuid(): string;
  setUuid(value: string): StoreWebFilesArtifactResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreWebFilesArtifactResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StoreWebFilesArtifactResponse): StoreWebFilesArtifactResponse.AsObject;
  static serializeBinaryToWriter(message: StoreWebFilesArtifactResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreWebFilesArtifactResponse;
  static deserializeBinaryFromReader(message: StoreWebFilesArtifactResponse, reader: jspb.BinaryReader): StoreWebFilesArtifactResponse;
}

export namespace StoreWebFilesArtifactResponse {
  export type AsObject = {
    uuid: string,
  }
}

export class StoreFilesArtifactFromServiceArgs extends jspb.Message {
  getServiceId(): string;
  setServiceId(value: string): StoreFilesArtifactFromServiceArgs;

  getSourcePath(): string;
  setSourcePath(value: string): StoreFilesArtifactFromServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreFilesArtifactFromServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StoreFilesArtifactFromServiceArgs): StoreFilesArtifactFromServiceArgs.AsObject;
  static serializeBinaryToWriter(message: StoreFilesArtifactFromServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreFilesArtifactFromServiceArgs;
  static deserializeBinaryFromReader(message: StoreFilesArtifactFromServiceArgs, reader: jspb.BinaryReader): StoreFilesArtifactFromServiceArgs;
}

export namespace StoreFilesArtifactFromServiceArgs {
  export type AsObject = {
    serviceId: string,
    sourcePath: string,
  }
}

export class StoreFilesArtifactFromServiceResponse extends jspb.Message {
  getUuid(): string;
  setUuid(value: string): StoreFilesArtifactFromServiceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreFilesArtifactFromServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StoreFilesArtifactFromServiceResponse): StoreFilesArtifactFromServiceResponse.AsObject;
  static serializeBinaryToWriter(message: StoreFilesArtifactFromServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreFilesArtifactFromServiceResponse;
  static deserializeBinaryFromReader(message: StoreFilesArtifactFromServiceResponse, reader: jspb.BinaryReader): StoreFilesArtifactFromServiceResponse;
}

export namespace StoreFilesArtifactFromServiceResponse {
  export type AsObject = {
    uuid: string,
  }
}

