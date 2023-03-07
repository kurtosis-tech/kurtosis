import * as jspb from 'google-protobuf'

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';
import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';


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
  getEnclaveName(): string;
  setEnclaveName(value: string): CreateEnclaveArgs;

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
    enclaveName: string,
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
  getEnclaveUuid(): string;
  setEnclaveUuid(value: string): EnclaveInfo;

  getName(): string;
  setName(value: string): EnclaveInfo;

  getShortenedUuid(): string;
  setShortenedUuid(value: string): EnclaveInfo;

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

  getCreationTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setCreationTime(value?: google_protobuf_timestamp_pb.Timestamp): EnclaveInfo;
  hasCreationTime(): boolean;
  clearCreationTime(): EnclaveInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveInfo.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveInfo): EnclaveInfo.AsObject;
  static serializeBinaryToWriter(message: EnclaveInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveInfo;
  static deserializeBinaryFromReader(message: EnclaveInfo, reader: jspb.BinaryReader): EnclaveInfo;
}

export namespace EnclaveInfo {
  export type AsObject = {
    enclaveUuid: string,
    name: string,
    shortenedUuid: string,
    containersStatus: EnclaveContainersStatus,
    apiContainerStatus: EnclaveAPIContainerStatus,
    apiContainerInfo?: EnclaveAPIContainerInfo.AsObject,
    apiContainerHostMachineInfo?: EnclaveAPIContainerHostMachineInfo.AsObject,
    creationTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
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

export class EnclaveIdentifiers extends jspb.Message {
  getEnclaveUuid(): string;
  setEnclaveUuid(value: string): EnclaveIdentifiers;

  getName(): string;
  setName(value: string): EnclaveIdentifiers;

  getShortenedUuid(): string;
  setShortenedUuid(value: string): EnclaveIdentifiers;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveIdentifiers.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveIdentifiers): EnclaveIdentifiers.AsObject;
  static serializeBinaryToWriter(message: EnclaveIdentifiers, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveIdentifiers;
  static deserializeBinaryFromReader(message: EnclaveIdentifiers, reader: jspb.BinaryReader): EnclaveIdentifiers;
}

export namespace EnclaveIdentifiers {
  export type AsObject = {
    enclaveUuid: string,
    name: string,
    shortenedUuid: string,
  }
}

export class GetExistingAndHistoricalEnclaveIdentifiersResponse extends jspb.Message {
  getAllidentifiersList(): Array<EnclaveIdentifiers>;
  setAllidentifiersList(value: Array<EnclaveIdentifiers>): GetExistingAndHistoricalEnclaveIdentifiersResponse;
  clearAllidentifiersList(): GetExistingAndHistoricalEnclaveIdentifiersResponse;
  addAllidentifiers(value?: EnclaveIdentifiers, index?: number): EnclaveIdentifiers;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetExistingAndHistoricalEnclaveIdentifiersResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetExistingAndHistoricalEnclaveIdentifiersResponse): GetExistingAndHistoricalEnclaveIdentifiersResponse.AsObject;
  static serializeBinaryToWriter(message: GetExistingAndHistoricalEnclaveIdentifiersResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetExistingAndHistoricalEnclaveIdentifiersResponse;
  static deserializeBinaryFromReader(message: GetExistingAndHistoricalEnclaveIdentifiersResponse, reader: jspb.BinaryReader): GetExistingAndHistoricalEnclaveIdentifiersResponse;
}

export namespace GetExistingAndHistoricalEnclaveIdentifiersResponse {
  export type AsObject = {
    allidentifiersList: Array<EnclaveIdentifiers.AsObject>,
  }
}

export class StopEnclaveArgs extends jspb.Message {
  getEnclaveIdentifier(): string;
  setEnclaveIdentifier(value: string): StopEnclaveArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StopEnclaveArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StopEnclaveArgs): StopEnclaveArgs.AsObject;
  static serializeBinaryToWriter(message: StopEnclaveArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StopEnclaveArgs;
  static deserializeBinaryFromReader(message: StopEnclaveArgs, reader: jspb.BinaryReader): StopEnclaveArgs;
}

export namespace StopEnclaveArgs {
  export type AsObject = {
    enclaveIdentifier: string,
  }
}

export class DestroyEnclaveArgs extends jspb.Message {
  getEnclaveIdentifier(): string;
  setEnclaveIdentifier(value: string): DestroyEnclaveArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DestroyEnclaveArgs.AsObject;
  static toObject(includeInstance: boolean, msg: DestroyEnclaveArgs): DestroyEnclaveArgs.AsObject;
  static serializeBinaryToWriter(message: DestroyEnclaveArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DestroyEnclaveArgs;
  static deserializeBinaryFromReader(message: DestroyEnclaveArgs, reader: jspb.BinaryReader): DestroyEnclaveArgs;
}

export namespace DestroyEnclaveArgs {
  export type AsObject = {
    enclaveIdentifier: string,
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

export class EnclaveNameAndUuid extends jspb.Message {
  getName(): string;
  setName(value: string): EnclaveNameAndUuid;

  getUuid(): string;
  setUuid(value: string): EnclaveNameAndUuid;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnclaveNameAndUuid.AsObject;
  static toObject(includeInstance: boolean, msg: EnclaveNameAndUuid): EnclaveNameAndUuid.AsObject;
  static serializeBinaryToWriter(message: EnclaveNameAndUuid, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnclaveNameAndUuid;
  static deserializeBinaryFromReader(message: EnclaveNameAndUuid, reader: jspb.BinaryReader): EnclaveNameAndUuid;
}

export namespace EnclaveNameAndUuid {
  export type AsObject = {
    name: string,
    uuid: string,
  }
}

export class CleanResponse extends jspb.Message {
  getRemovedEnclaveNameAndUuidsList(): Array<EnclaveNameAndUuid>;
  setRemovedEnclaveNameAndUuidsList(value: Array<EnclaveNameAndUuid>): CleanResponse;
  clearRemovedEnclaveNameAndUuidsList(): CleanResponse;
  addRemovedEnclaveNameAndUuids(value?: EnclaveNameAndUuid, index?: number): EnclaveNameAndUuid;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CleanResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CleanResponse): CleanResponse.AsObject;
  static serializeBinaryToWriter(message: CleanResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CleanResponse;
  static deserializeBinaryFromReader(message: CleanResponse, reader: jspb.BinaryReader): CleanResponse;
}

export namespace CleanResponse {
  export type AsObject = {
    removedEnclaveNameAndUuidsList: Array<EnclaveNameAndUuid.AsObject>,
  }
}

export class GetServiceLogsArgs extends jspb.Message {
  getEnclaveIdentifier(): string;
  setEnclaveIdentifier(value: string): GetServiceLogsArgs;

  getServiceUuidSetMap(): jspb.Map<string, boolean>;
  clearServiceUuidSetMap(): GetServiceLogsArgs;

  getFollowLogs(): boolean;
  setFollowLogs(value: boolean): GetServiceLogsArgs;

  getConjunctiveFiltersList(): Array<LogLineFilter>;
  setConjunctiveFiltersList(value: Array<LogLineFilter>): GetServiceLogsArgs;
  clearConjunctiveFiltersList(): GetServiceLogsArgs;
  addConjunctiveFilters(value?: LogLineFilter, index?: number): LogLineFilter;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceLogsArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceLogsArgs): GetServiceLogsArgs.AsObject;
  static serializeBinaryToWriter(message: GetServiceLogsArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServiceLogsArgs;
  static deserializeBinaryFromReader(message: GetServiceLogsArgs, reader: jspb.BinaryReader): GetServiceLogsArgs;
}

export namespace GetServiceLogsArgs {
  export type AsObject = {
    enclaveIdentifier: string,
    serviceUuidSetMap: Array<[string, boolean]>,
    followLogs: boolean,
    conjunctiveFiltersList: Array<LogLineFilter.AsObject>,
  }
}

export class GetServiceLogsResponse extends jspb.Message {
  getServiceLogsByServiceUuidMap(): jspb.Map<string, LogLine>;
  clearServiceLogsByServiceUuidMap(): GetServiceLogsResponse;

  getNotFoundServiceUuidSetMap(): jspb.Map<string, boolean>;
  clearNotFoundServiceUuidSetMap(): GetServiceLogsResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServiceLogsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServiceLogsResponse): GetServiceLogsResponse.AsObject;
  static serializeBinaryToWriter(message: GetServiceLogsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServiceLogsResponse;
  static deserializeBinaryFromReader(message: GetServiceLogsResponse, reader: jspb.BinaryReader): GetServiceLogsResponse;
}

export namespace GetServiceLogsResponse {
  export type AsObject = {
    serviceLogsByServiceUuidMap: Array<[string, LogLine.AsObject]>,
    notFoundServiceUuidSetMap: Array<[string, boolean]>,
  }
}

export class LogLine extends jspb.Message {
  getLineList(): Array<string>;
  setLineList(value: Array<string>): LogLine;
  clearLineList(): LogLine;
  addLine(value: string, index?: number): LogLine;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LogLine.AsObject;
  static toObject(includeInstance: boolean, msg: LogLine): LogLine.AsObject;
  static serializeBinaryToWriter(message: LogLine, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LogLine;
  static deserializeBinaryFromReader(message: LogLine, reader: jspb.BinaryReader): LogLine;
}

export namespace LogLine {
  export type AsObject = {
    lineList: Array<string>,
  }
}

export class LogLineFilter extends jspb.Message {
  getOperator(): LogLineOperator;
  setOperator(value: LogLineOperator): LogLineFilter;

  getTextPattern(): string;
  setTextPattern(value: string): LogLineFilter;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LogLineFilter.AsObject;
  static toObject(includeInstance: boolean, msg: LogLineFilter): LogLineFilter.AsObject;
  static serializeBinaryToWriter(message: LogLineFilter, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LogLineFilter;
  static deserializeBinaryFromReader(message: LogLineFilter, reader: jspb.BinaryReader): LogLineFilter;
}

export namespace LogLineFilter {
  export type AsObject = {
    operator: LogLineOperator,
    textPattern: string,
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
export enum LogLineOperator { 
  LOGLINEOPERATOR_DOES_CONTAIN_TEXT = 0,
  LOGLINEOPERATOR_DOES_NOT_CONTAIN_TEXT = 1,
  LOGLINEOPERATOR_DOES_CONTAIN_MATCH_REGEX = 2,
  LOGLINEOPERATOR_DOES_NOT_CONTAIN_MATCH_REGEX = 3,
}
