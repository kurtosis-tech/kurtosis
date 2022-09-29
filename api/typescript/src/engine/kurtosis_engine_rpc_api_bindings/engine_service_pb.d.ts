import * as jspb from 'google-protobuf'

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class GetEngineInfoResponse extends jspb.Message {
  getEngineVersion(): string;
  setEngineVersion(value: string): GetEngineInfoResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetEngineInfoResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetEngineInfoResponse): GetEngineInfoResponse.AsObject;
  static serializeBinaryToWriter(message: GetEngineInfoResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetEngineInfoResponse;
  static deserializeBinaryFromReader(message: GetEngineInfoResponse, reader: jspb.BinaryReader): GetEngineInfoResponse;
}

export namespace GetEngineInfoResponse {
  export type AsObject = {
    engineVersion: string,
  }
}

export class CreateEnclaveArgs extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): CreateEnclaveArgs;

  getApiContainerVersionTag(): string;
  setApiContainerVersionTag(value: string): CreateEnclaveArgs;

  getApiContainerLogLevel(): string;
  setApiContainerLogLevel(value: string): CreateEnclaveArgs;

  getIsPartitioningEnabled(): boolean;
  setIsPartitioningEnabled(value: boolean): CreateEnclaveArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateEnclaveArgs.AsObject;
  static toObject(includeInstance: boolean, msg: CreateEnclaveArgs): CreateEnclaveArgs.AsObject;
  static serializeBinaryToWriter(message: CreateEnclaveArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateEnclaveArgs;
  static deserializeBinaryFromReader(message: CreateEnclaveArgs, reader: jspb.BinaryReader): CreateEnclaveArgs;
}

export namespace CreateEnclaveArgs {
  export type AsObject = {
    enclaveId: string,
    apiContainerVersionTag: string,
    apiContainerLogLevel: string,
    isPartitioningEnabled: boolean,
  }
}

export class CreateEnclaveResponse extends jspb.Message {
  getEnclaveInfo(): EnclaveInfo | undefined;
  setEnclaveInfo(value?: EnclaveInfo): CreateEnclaveResponse;
  hasEnclaveInfo(): boolean;
  clearEnclaveInfo(): CreateEnclaveResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateEnclaveResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateEnclaveResponse): CreateEnclaveResponse.AsObject;
  static serializeBinaryToWriter(message: CreateEnclaveResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateEnclaveResponse;
  static deserializeBinaryFromReader(message: CreateEnclaveResponse, reader: jspb.BinaryReader): CreateEnclaveResponse;
}

export namespace CreateEnclaveResponse {
  export type AsObject = {
    enclaveInfo?: EnclaveInfo.AsObject,
  }
}

export class EnclaveAPIContainerInfo extends jspb.Message {
  getContainerId(): string;
  setContainerId(value: string): EnclaveAPIContainerInfo;

  getIpInsideEnclave(): string;
  setIpInsideEnclave(value: string): EnclaveAPIContainerInfo;

  getGrpcPortInsideEnclave(): number;
  setGrpcPortInsideEnclave(value: number): EnclaveAPIContainerInfo;

  getGrpcProxyPortInsideEnclave(): number;
  setGrpcProxyPortInsideEnclave(value: number): EnclaveAPIContainerInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveAPIContainerInfo.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveAPIContainerInfo): EnclaveAPIContainerInfo.AsObject;
  static serializeBinaryToWriter(message: EnclaveAPIContainerInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveAPIContainerInfo;
  static deserializeBinaryFromReader(message: EnclaveAPIContainerInfo, reader: jspb.BinaryReader): EnclaveAPIContainerInfo;
}

export namespace EnclaveAPIContainerInfo {
  export type AsObject = {
    containerId: string,
    ipInsideEnclave: string,
    grpcPortInsideEnclave: number,
    grpcProxyPortInsideEnclave: number,
  }
}

export class EnclaveAPIContainerHostMachineInfo extends jspb.Message {
  getIpOnHostMachine(): string;
  setIpOnHostMachine(value: string): EnclaveAPIContainerHostMachineInfo;

  getGrpcPortOnHostMachine(): number;
  setGrpcPortOnHostMachine(value: number): EnclaveAPIContainerHostMachineInfo;

  getGrpcProxyPortOnHostMachine(): number;
  setGrpcProxyPortOnHostMachine(value: number): EnclaveAPIContainerHostMachineInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveAPIContainerHostMachineInfo.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveAPIContainerHostMachineInfo): EnclaveAPIContainerHostMachineInfo.AsObject;
  static serializeBinaryToWriter(message: EnclaveAPIContainerHostMachineInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveAPIContainerHostMachineInfo;
  static deserializeBinaryFromReader(message: EnclaveAPIContainerHostMachineInfo, reader: jspb.BinaryReader): EnclaveAPIContainerHostMachineInfo;
}

export namespace EnclaveAPIContainerHostMachineInfo {
  export type AsObject = {
    ipOnHostMachine: string,
    grpcPortOnHostMachine: number,
    grpcProxyPortOnHostMachine: number,
  }
}

export class EnclaveInfo extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): EnclaveInfo;

  getContainersStatus(): EnclaveContainersStatus;
  setContainersStatus(value: EnclaveContainersStatus): EnclaveInfo;

  getApiContainerStatus(): EnclaveAPIContainerStatus;
  setApiContainerStatus(value: EnclaveAPIContainerStatus): EnclaveInfo;

  getApiContainerInfo(): EnclaveAPIContainerInfo | undefined;
  setApiContainerInfo(value?: EnclaveAPIContainerInfo): EnclaveInfo;
  hasApiContainerInfo(): boolean;
  clearApiContainerInfo(): EnclaveInfo;

  getApiContainerHostMachineInfo(): EnclaveAPIContainerHostMachineInfo | undefined;
  setApiContainerHostMachineInfo(value?: EnclaveAPIContainerHostMachineInfo): EnclaveInfo;
  hasApiContainerHostMachineInfo(): boolean;
  clearApiContainerHostMachineInfo(): EnclaveInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveInfo.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveInfo): EnclaveInfo.AsObject;
  static serializeBinaryToWriter(message: EnclaveInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveInfo;
  static deserializeBinaryFromReader(message: EnclaveInfo, reader: jspb.BinaryReader): EnclaveInfo;
}

export namespace EnclaveInfo {
  export type AsObject = {
    enclaveId: string,
    containersStatus: EnclaveContainersStatus,
    apiContainerStatus: EnclaveAPIContainerStatus,
    apiContainerInfo?: EnclaveAPIContainerInfo.AsObject,
    apiContainerHostMachineInfo?: EnclaveAPIContainerHostMachineInfo.AsObject,
  }
}

export class GetEnclavesResponse extends jspb.Message {
  getEnclaveInfoMap(): jspb.Map<string, EnclaveInfo>;
  clearEnclaveInfoMap(): GetEnclavesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetEnclavesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetEnclavesResponse): GetEnclavesResponse.AsObject;
  static serializeBinaryToWriter(message: GetEnclavesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetEnclavesResponse;
  static deserializeBinaryFromReader(message: GetEnclavesResponse, reader: jspb.BinaryReader): GetEnclavesResponse;
}

export namespace GetEnclavesResponse {
  export type AsObject = {
    enclaveInfoMap: Array<[string, EnclaveInfo.AsObject]>,
  }
}

export class StopEnclaveArgs extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): StopEnclaveArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StopEnclaveArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StopEnclaveArgs): StopEnclaveArgs.AsObject;
  static serializeBinaryToWriter(message: StopEnclaveArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StopEnclaveArgs;
  static deserializeBinaryFromReader(message: StopEnclaveArgs, reader: jspb.BinaryReader): StopEnclaveArgs;
}

export namespace StopEnclaveArgs {
  export type AsObject = {
    enclaveId: string,
  }
}

export class DestroyEnclaveArgs extends jspb.Message {
  getEnclaveId(): string;
  setEnclaveId(value: string): DestroyEnclaveArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DestroyEnclaveArgs.AsObject;
  static toObject(includeInstance: boolean, msg: DestroyEnclaveArgs): DestroyEnclaveArgs.AsObject;
  static serializeBinaryToWriter(message: DestroyEnclaveArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DestroyEnclaveArgs;
  static deserializeBinaryFromReader(message: DestroyEnclaveArgs, reader: jspb.BinaryReader): DestroyEnclaveArgs;
}

export namespace DestroyEnclaveArgs {
  export type AsObject = {
    enclaveId: string,
  }
}

export class CleanArgs extends jspb.Message {
  getShouldCleanAll(): boolean;
  setShouldCleanAll(value: boolean): CleanArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CleanArgs.AsObject;
  static toObject(includeInstance: boolean, msg: CleanArgs): CleanArgs.AsObject;
  static serializeBinaryToWriter(message: CleanArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CleanArgs;
  static deserializeBinaryFromReader(message: CleanArgs, reader: jspb.BinaryReader): CleanArgs;
}

export namespace CleanArgs {
  export type AsObject = {
    shouldCleanAll: boolean,
  }
}

export class CleanResponse extends jspb.Message {
  getRemovedEnclaveIdsMap(): jspb.Map<string, boolean>;
  clearRemovedEnclaveIdsMap(): CleanResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CleanResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CleanResponse): CleanResponse.AsObject;
  static serializeBinaryToWriter(message: CleanResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CleanResponse;
  static deserializeBinaryFromReader(message: CleanResponse, reader: jspb.BinaryReader): CleanResponse;
}

export namespace CleanResponse {
  export type AsObject = {
    removedEnclaveIdsMap: Array<[string, boolean]>,
  }
}

export enum EnclaveContainersStatus { 
  ENCLAVECONTAINERSSTATUS_EMPTY = 0,
  ENCLAVECONTAINERSSTATUS_RUNNING = 1,
  ENCLAVECONTAINERSSTATUS_STOPPED = 2,
}
export enum EnclaveAPIContainerStatus { 
  ENCLAVEAPICONTAINERSTATUS_NONEXISTENT = 0,
  ENCLAVEAPICONTAINERSTATUS_RUNNING = 1,
  ENCLAVEAPICONTAINERSTATUS_STOPPED = 2,
}
