import * as jspb from 'google-protobuf'

import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';


export class Port extends jspb.Message {
  getNumber(): number;
  setNumber(value: number): Port;

  getTransportProtocol(): Port.TransportProtocol;
  setTransportProtocol(value: Port.TransportProtocol): Port;

  getMaybeApplicationProtocol(): string;
  setMaybeApplicationProtocol(value: string): Port;

  getMaybeWaitTimeout(): string;
  setMaybeWaitTimeout(value: string): Port;

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
    transportProtocol: Port.TransportProtocol,
    maybeApplicationProtocol: string,
    maybeWaitTimeout: string,
  }

  export enum TransportProtocol { 
    TCP = 0,
    SCTP = 1,
    UDP = 2,
  }
}

export class ServiceInfo extends jspb.Message {
  getServiceUuid(): string;
  setServiceUuid(value: string): ServiceInfo;

  getPrivateIpAddr(): string;
  setPrivateIpAddr(value: string): ServiceInfo;

  getPrivatePortsMap(): jspb.Map<string, Port>;
  clearPrivatePortsMap(): ServiceInfo;

  getMaybePublicIpAddr(): string;
  setMaybePublicIpAddr(value: string): ServiceInfo;

  getMaybePublicPortsMap(): jspb.Map<string, Port>;
  clearMaybePublicPortsMap(): ServiceInfo;

  getName(): string;
  setName(value: string): ServiceInfo;

  getShortenedUuid(): string;
  setShortenedUuid(value: string): ServiceInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServiceInfo.AsObject;
  static toObject(includeInstance: boolean, msg: ServiceInfo): ServiceInfo.AsObject;
  static serializeBinaryToWriter(message: ServiceInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServiceInfo;
  static deserializeBinaryFromReader(message: ServiceInfo, reader: jspb.BinaryReader): ServiceInfo;
}

export namespace ServiceInfo {
  export type AsObject = {
    serviceUuid: string,
    privateIpAddr: string,
    privatePortsMap: Array<[string, Port.AsObject]>,
    maybePublicIpAddr: string,
    maybePublicPortsMap: Array<[string, Port.AsObject]>,
    name: string,
    shortenedUuid: string,
  }
}

export class ServiceConfig extends jspb.Message {
  getContainerImageName(): string;
  setContainerImageName(value: string): ServiceConfig;

  getPrivatePortsMap(): jspb.Map<string, Port>;
  clearPrivatePortsMap(): ServiceConfig;

  getPublicPortsMap(): jspb.Map<string, Port>;
  clearPublicPortsMap(): ServiceConfig;

  getEntrypointArgsList(): Array<string>;
  setEntrypointArgsList(value: Array<string>): ServiceConfig;
  clearEntrypointArgsList(): ServiceConfig;
  addEntrypointArgs(value: string, index?: number): ServiceConfig;

  getCmdArgsList(): Array<string>;
  setCmdArgsList(value: Array<string>): ServiceConfig;
  clearCmdArgsList(): ServiceConfig;
  addCmdArgs(value: string, index?: number): ServiceConfig;

  getEnvVarsMap(): jspb.Map<string, string>;
  clearEnvVarsMap(): ServiceConfig;

  getFilesArtifactMountpointsMap(): jspb.Map<string, string>;
  clearFilesArtifactMountpointsMap(): ServiceConfig;

  getCpuAllocationMillicpus(): number;
  setCpuAllocationMillicpus(value: number): ServiceConfig;

  getMemoryAllocationMegabytes(): number;
  setMemoryAllocationMegabytes(value: number): ServiceConfig;

  getPrivateIpAddrPlaceholder(): string;
  setPrivateIpAddrPlaceholder(value: string): ServiceConfig;

  getSubnetwork(): string;
  setSubnetwork(value: string): ServiceConfig;
  hasSubnetwork(): boolean;
  clearSubnetwork(): ServiceConfig;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServiceConfig.AsObject;
  static toObject(includeInstance: boolean, msg: ServiceConfig): ServiceConfig.AsObject;
  static serializeBinaryToWriter(message: ServiceConfig, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServiceConfig;
  static deserializeBinaryFromReader(message: ServiceConfig, reader: jspb.BinaryReader): ServiceConfig;
}

export namespace ServiceConfig {
  export type AsObject = {
    containerImageName: string,
    privatePortsMap: Array<[string, Port.AsObject]>,
    publicPortsMap: Array<[string, Port.AsObject]>,
    entrypointArgsList: Array<string>,
    cmdArgsList: Array<string>,
    envVarsMap: Array<[string, string]>,
    filesArtifactMountpointsMap: Array<[string, string]>,
    cpuAllocationMillicpus: number,
    memoryAllocationMegabytes: number,
    privateIpAddrPlaceholder: string,
    subnetwork?: string,
  }

  export enum SubnetworkCase { 
    _SUBNETWORK_NOT_SET = 0,
    SUBNETWORK = 11,
  }
}

export class UpdateServiceConfig extends jspb.Message {
  getSubnetwork(): string;
  setSubnetwork(value: string): UpdateServiceConfig;
  hasSubnetwork(): boolean;
  clearSubnetwork(): UpdateServiceConfig;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateServiceConfig.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateServiceConfig): UpdateServiceConfig.AsObject;
  static serializeBinaryToWriter(message: UpdateServiceConfig, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateServiceConfig;
  static deserializeBinaryFromReader(message: UpdateServiceConfig, reader: jspb.BinaryReader): UpdateServiceConfig;
}

export namespace UpdateServiceConfig {
  export type AsObject = {
    subnetwork?: string,
  }

  export enum SubnetworkCase { 
    _SUBNETWORK_NOT_SET = 0,
    SUBNETWORK = 1,
  }
}

export class RunStarlarkScriptArgs extends jspb.Message {
  getSerializedScript(): string;
  setSerializedScript(value: string): RunStarlarkScriptArgs;

  getSerializedParams(): string;
  setSerializedParams(value: string): RunStarlarkScriptArgs;

  getDryRun(): boolean;
  setDryRun(value: boolean): RunStarlarkScriptArgs;
  hasDryRun(): boolean;
  clearDryRun(): RunStarlarkScriptArgs;

  getParallelism(): number;
  setParallelism(value: number): RunStarlarkScriptArgs;
  hasParallelism(): boolean;
  clearParallelism(): RunStarlarkScriptArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RunStarlarkScriptArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RunStarlarkScriptArgs): RunStarlarkScriptArgs.AsObject;
  static serializeBinaryToWriter(message: RunStarlarkScriptArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RunStarlarkScriptArgs;
  static deserializeBinaryFromReader(message: RunStarlarkScriptArgs, reader: jspb.BinaryReader): RunStarlarkScriptArgs;
}

export namespace RunStarlarkScriptArgs {
  export type AsObject = {
    serializedScript: string,
    serializedParams: string,
    dryRun?: boolean,
    parallelism?: number,
  }

  export enum DryRunCase { 
    _DRY_RUN_NOT_SET = 0,
    DRY_RUN = 3,
  }

  export enum ParallelismCase { 
    _PARALLELISM_NOT_SET = 0,
    PARALLELISM = 4,
  }
}

export class RunStarlarkPackageArgs extends jspb.Message {
  getPackageId(): string;
  setPackageId(value: string): RunStarlarkPackageArgs;

  getLocal(): Uint8Array | string;
  getLocal_asU8(): Uint8Array;
  getLocal_asB64(): string;
  setLocal(value: Uint8Array | string): RunStarlarkPackageArgs;

  getRemote(): boolean;
  setRemote(value: boolean): RunStarlarkPackageArgs;

  getSerializedParams(): string;
  setSerializedParams(value: string): RunStarlarkPackageArgs;

  getDryRun(): boolean;
  setDryRun(value: boolean): RunStarlarkPackageArgs;
  hasDryRun(): boolean;
  clearDryRun(): RunStarlarkPackageArgs;

  getParallelism(): number;
  setParallelism(value: number): RunStarlarkPackageArgs;
  hasParallelism(): boolean;
  clearParallelism(): RunStarlarkPackageArgs;

  getClonePackage(): boolean;
  setClonePackage(value: boolean): RunStarlarkPackageArgs;
  hasClonePackage(): boolean;
  clearClonePackage(): RunStarlarkPackageArgs;

  getStarlarkPackageContentCase(): RunStarlarkPackageArgs.StarlarkPackageContentCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RunStarlarkPackageArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RunStarlarkPackageArgs): RunStarlarkPackageArgs.AsObject;
  static serializeBinaryToWriter(message: RunStarlarkPackageArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RunStarlarkPackageArgs;
  static deserializeBinaryFromReader(message: RunStarlarkPackageArgs, reader: jspb.BinaryReader): RunStarlarkPackageArgs;
}

export namespace RunStarlarkPackageArgs {
  export type AsObject = {
    packageId: string,
    local: Uint8Array | string,
    remote: boolean,
    serializedParams: string,
    dryRun?: boolean,
    parallelism?: number,
    clonePackage?: boolean,
  }

  export enum StarlarkPackageContentCase { 
    STARLARK_PACKAGE_CONTENT_NOT_SET = 0,
    LOCAL = 3,
    REMOTE = 4,
  }

  export enum DryRunCase { 
    _DRY_RUN_NOT_SET = 0,
    DRY_RUN = 6,
  }

  export enum ParallelismCase { 
    _PARALLELISM_NOT_SET = 0,
    PARALLELISM = 7,
  }

  export enum ClonePackageCase { 
    _CLONE_PACKAGE_NOT_SET = 0,
    CLONE_PACKAGE = 8,
  }
}

export class StarlarkRunResponseLine extends jspb.Message {
  getInstruction(): StarlarkInstruction | undefined;
  setInstruction(value?: StarlarkInstruction): StarlarkRunResponseLine;
  hasInstruction(): boolean;
  clearInstruction(): StarlarkRunResponseLine;

  getError(): StarlarkError | undefined;
  setError(value?: StarlarkError): StarlarkRunResponseLine;
  hasError(): boolean;
  clearError(): StarlarkRunResponseLine;

  getProgressInfo(): StarlarkRunProgress | undefined;
  setProgressInfo(value?: StarlarkRunProgress): StarlarkRunResponseLine;
  hasProgressInfo(): boolean;
  clearProgressInfo(): StarlarkRunResponseLine;

  getInstructionResult(): StarlarkInstructionResult | undefined;
  setInstructionResult(value?: StarlarkInstructionResult): StarlarkRunResponseLine;
  hasInstructionResult(): boolean;
  clearInstructionResult(): StarlarkRunResponseLine;

  getRunFinishedEvent(): StarlarkRunFinishedEvent | undefined;
  setRunFinishedEvent(value?: StarlarkRunFinishedEvent): StarlarkRunResponseLine;
  hasRunFinishedEvent(): boolean;
  clearRunFinishedEvent(): StarlarkRunResponseLine;

  getWarning(): StarlarkWarning | undefined;
  setWarning(value?: StarlarkWarning): StarlarkRunResponseLine;
  hasWarning(): boolean;
  clearWarning(): StarlarkRunResponseLine;

  getRunResponseLineCase(): StarlarkRunResponseLine.RunResponseLineCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkRunResponseLine.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkRunResponseLine): StarlarkRunResponseLine.AsObject;
  static serializeBinaryToWriter(message: StarlarkRunResponseLine, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkRunResponseLine;
  static deserializeBinaryFromReader(message: StarlarkRunResponseLine, reader: jspb.BinaryReader): StarlarkRunResponseLine;
}

export namespace StarlarkRunResponseLine {
  export type AsObject = {
    instruction?: StarlarkInstruction.AsObject,
    error?: StarlarkError.AsObject,
    progressInfo?: StarlarkRunProgress.AsObject,
    instructionResult?: StarlarkInstructionResult.AsObject,
    runFinishedEvent?: StarlarkRunFinishedEvent.AsObject,
    warning?: StarlarkWarning.AsObject,
  }

  export enum RunResponseLineCase { 
    RUN_RESPONSE_LINE_NOT_SET = 0,
    INSTRUCTION = 1,
    ERROR = 2,
    PROGRESS_INFO = 3,
    INSTRUCTION_RESULT = 4,
    RUN_FINISHED_EVENT = 5,
    WARNING = 6,
  }
}

export class StarlarkWarning extends jspb.Message {
  getWarningMessage(): string;
  setWarningMessage(value: string): StarlarkWarning;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkWarning.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkWarning): StarlarkWarning.AsObject;
  static serializeBinaryToWriter(message: StarlarkWarning, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkWarning;
  static deserializeBinaryFromReader(message: StarlarkWarning, reader: jspb.BinaryReader): StarlarkWarning;
}

export namespace StarlarkWarning {
  export type AsObject = {
    warningMessage: string,
  }
}

export class StarlarkInstruction extends jspb.Message {
  getPosition(): StarlarkInstructionPosition | undefined;
  setPosition(value?: StarlarkInstructionPosition): StarlarkInstruction;
  hasPosition(): boolean;
  clearPosition(): StarlarkInstruction;

  getInstructionName(): string;
  setInstructionName(value: string): StarlarkInstruction;

  getArgumentsList(): Array<StarlarkInstructionArg>;
  setArgumentsList(value: Array<StarlarkInstructionArg>): StarlarkInstruction;
  clearArgumentsList(): StarlarkInstruction;
  addArguments(value?: StarlarkInstructionArg, index?: number): StarlarkInstructionArg;

  getExecutableInstruction(): string;
  setExecutableInstruction(value: string): StarlarkInstruction;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkInstruction.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkInstruction): StarlarkInstruction.AsObject;
  static serializeBinaryToWriter(message: StarlarkInstruction, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkInstruction;
  static deserializeBinaryFromReader(message: StarlarkInstruction, reader: jspb.BinaryReader): StarlarkInstruction;
}

export namespace StarlarkInstruction {
  export type AsObject = {
    position?: StarlarkInstructionPosition.AsObject,
    instructionName: string,
    argumentsList: Array<StarlarkInstructionArg.AsObject>,
    executableInstruction: string,
  }
}

export class StarlarkInstructionResult extends jspb.Message {
  getSerializedInstructionResult(): string;
  setSerializedInstructionResult(value: string): StarlarkInstructionResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkInstructionResult.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkInstructionResult): StarlarkInstructionResult.AsObject;
  static serializeBinaryToWriter(message: StarlarkInstructionResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkInstructionResult;
  static deserializeBinaryFromReader(message: StarlarkInstructionResult, reader: jspb.BinaryReader): StarlarkInstructionResult;
}

export namespace StarlarkInstructionResult {
  export type AsObject = {
    serializedInstructionResult: string,
  }
}

export class StarlarkInstructionArg extends jspb.Message {
  getSerializedArgValue(): string;
  setSerializedArgValue(value: string): StarlarkInstructionArg;

  getArgName(): string;
  setArgName(value: string): StarlarkInstructionArg;
  hasArgName(): boolean;
  clearArgName(): StarlarkInstructionArg;

  getIsRepresentative(): boolean;
  setIsRepresentative(value: boolean): StarlarkInstructionArg;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkInstructionArg.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkInstructionArg): StarlarkInstructionArg.AsObject;
  static serializeBinaryToWriter(message: StarlarkInstructionArg, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkInstructionArg;
  static deserializeBinaryFromReader(message: StarlarkInstructionArg, reader: jspb.BinaryReader): StarlarkInstructionArg;
}

export namespace StarlarkInstructionArg {
  export type AsObject = {
    serializedArgValue: string,
    argName?: string,
    isRepresentative: boolean,
  }

  export enum ArgNameCase { 
    _ARG_NAME_NOT_SET = 0,
    ARG_NAME = 2,
  }
}

export class StarlarkInstructionPosition extends jspb.Message {
  getFilename(): string;
  setFilename(value: string): StarlarkInstructionPosition;

  getLine(): number;
  setLine(value: number): StarlarkInstructionPosition;

  getColumn(): number;
  setColumn(value: number): StarlarkInstructionPosition;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkInstructionPosition.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkInstructionPosition): StarlarkInstructionPosition.AsObject;
  static serializeBinaryToWriter(message: StarlarkInstructionPosition, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkInstructionPosition;
  static deserializeBinaryFromReader(message: StarlarkInstructionPosition, reader: jspb.BinaryReader): StarlarkInstructionPosition;
}

export namespace StarlarkInstructionPosition {
  export type AsObject = {
    filename: string,
    line: number,
    column: number,
  }
}

export class StarlarkError extends jspb.Message {
  getInterpretationError(): StarlarkInterpretationError | undefined;
  setInterpretationError(value?: StarlarkInterpretationError): StarlarkError;
  hasInterpretationError(): boolean;
  clearInterpretationError(): StarlarkError;

  getValidationError(): StarlarkValidationError | undefined;
  setValidationError(value?: StarlarkValidationError): StarlarkError;
  hasValidationError(): boolean;
  clearValidationError(): StarlarkError;

  getExecutionError(): StarlarkExecutionError | undefined;
  setExecutionError(value?: StarlarkExecutionError): StarlarkError;
  hasExecutionError(): boolean;
  clearExecutionError(): StarlarkError;

  getErrorCase(): StarlarkError.ErrorCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkError.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkError): StarlarkError.AsObject;
  static serializeBinaryToWriter(message: StarlarkError, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkError;
  static deserializeBinaryFromReader(message: StarlarkError, reader: jspb.BinaryReader): StarlarkError;
}

export namespace StarlarkError {
  export type AsObject = {
    interpretationError?: StarlarkInterpretationError.AsObject,
    validationError?: StarlarkValidationError.AsObject,
    executionError?: StarlarkExecutionError.AsObject,
  }

  export enum ErrorCase { 
    ERROR_NOT_SET = 0,
    INTERPRETATION_ERROR = 1,
    VALIDATION_ERROR = 2,
    EXECUTION_ERROR = 3,
  }
}

export class StarlarkInterpretationError extends jspb.Message {
  getErrorMessage(): string;
  setErrorMessage(value: string): StarlarkInterpretationError;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkInterpretationError.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkInterpretationError): StarlarkInterpretationError.AsObject;
  static serializeBinaryToWriter(message: StarlarkInterpretationError, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkInterpretationError;
  static deserializeBinaryFromReader(message: StarlarkInterpretationError, reader: jspb.BinaryReader): StarlarkInterpretationError;
}

export namespace StarlarkInterpretationError {
  export type AsObject = {
    errorMessage: string,
  }
}

export class StarlarkValidationError extends jspb.Message {
  getErrorMessage(): string;
  setErrorMessage(value: string): StarlarkValidationError;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkValidationError.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkValidationError): StarlarkValidationError.AsObject;
  static serializeBinaryToWriter(message: StarlarkValidationError, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkValidationError;
  static deserializeBinaryFromReader(message: StarlarkValidationError, reader: jspb.BinaryReader): StarlarkValidationError;
}

export namespace StarlarkValidationError {
  export type AsObject = {
    errorMessage: string,
  }
}

export class StarlarkExecutionError extends jspb.Message {
  getErrorMessage(): string;
  setErrorMessage(value: string): StarlarkExecutionError;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkExecutionError.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkExecutionError): StarlarkExecutionError.AsObject;
  static serializeBinaryToWriter(message: StarlarkExecutionError, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkExecutionError;
  static deserializeBinaryFromReader(message: StarlarkExecutionError, reader: jspb.BinaryReader): StarlarkExecutionError;
}

export namespace StarlarkExecutionError {
  export type AsObject = {
    errorMessage: string,
  }
}

export class StarlarkRunProgress extends jspb.Message {
  getCurrentStepInfoList(): Array<string>;
  setCurrentStepInfoList(value: Array<string>): StarlarkRunProgress;
  clearCurrentStepInfoList(): StarlarkRunProgress;
  addCurrentStepInfo(value: string, index?: number): StarlarkRunProgress;

  getTotalSteps(): number;
  setTotalSteps(value: number): StarlarkRunProgress;

  getCurrentStepNumber(): number;
  setCurrentStepNumber(value: number): StarlarkRunProgress;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkRunProgress.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkRunProgress): StarlarkRunProgress.AsObject;
  static serializeBinaryToWriter(message: StarlarkRunProgress, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkRunProgress;
  static deserializeBinaryFromReader(message: StarlarkRunProgress, reader: jspb.BinaryReader): StarlarkRunProgress;
}

export namespace StarlarkRunProgress {
  export type AsObject = {
    currentStepInfoList: Array<string>,
    totalSteps: number,
    currentStepNumber: number,
  }
}

export class StarlarkRunFinishedEvent extends jspb.Message {
  getIsrunsuccessful(): boolean;
  setIsrunsuccessful(value: boolean): StarlarkRunFinishedEvent;

  getSerializedOutput(): string;
  setSerializedOutput(value: string): StarlarkRunFinishedEvent;
  hasSerializedOutput(): boolean;
  clearSerializedOutput(): StarlarkRunFinishedEvent;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StarlarkRunFinishedEvent.AsObject;
  static toObject(includeInstance: boolean, msg: StarlarkRunFinishedEvent): StarlarkRunFinishedEvent.AsObject;
  static serializeBinaryToWriter(message: StarlarkRunFinishedEvent, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StarlarkRunFinishedEvent;
  static deserializeBinaryFromReader(message: StarlarkRunFinishedEvent, reader: jspb.BinaryReader): StarlarkRunFinishedEvent;
}

export namespace StarlarkRunFinishedEvent {
  export type AsObject = {
    isrunsuccessful: boolean,
    serializedOutput?: string,
  }

  export enum SerializedOutputCase { 
    _SERIALIZED_OUTPUT_NOT_SET = 0,
    SERIALIZED_OUTPUT = 2,
  }
}

export class StartServicesArgs extends jspb.Message {
  getServiceNamesToConfigsMap(): jspb.Map<string, ServiceConfig>;
  clearServiceNamesToConfigsMap(): StartServicesArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServicesArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StartServicesArgs): StartServicesArgs.AsObject;
  static serializeBinaryToWriter(message: StartServicesArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServicesArgs;
  static deserializeBinaryFromReader(message: StartServicesArgs, reader: jspb.BinaryReader): StartServicesArgs;
}

export namespace StartServicesArgs {
  export type AsObject = {
    serviceNamesToConfigsMap: Array<[string, ServiceConfig.AsObject]>,
  }
}

export class StartServicesResponse extends jspb.Message {
  getSuccessfulServiceNameToServiceInfoMap(): jspb.Map<string, ServiceInfo>;
  clearSuccessfulServiceNameToServiceInfoMap(): StartServicesResponse;

  getFailedServiceNameToErrorMap(): jspb.Map<string, string>;
  clearFailedServiceNameToErrorMap(): StartServicesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StartServicesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StartServicesResponse): StartServicesResponse.AsObject;
  static serializeBinaryToWriter(message: StartServicesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StartServicesResponse;
  static deserializeBinaryFromReader(message: StartServicesResponse, reader: jspb.BinaryReader): StartServicesResponse;
}

export namespace StartServicesResponse {
  export type AsObject = {
    successfulServiceNameToServiceInfoMap: Array<[string, ServiceInfo.AsObject]>,
    failedServiceNameToErrorMap: Array<[string, string]>,
  }
}

export class GetServicesArgs extends jspb.Message {
  getServiceIdentifiersMap(): jspb.Map<string, boolean>;
  clearServiceIdentifiersMap(): GetServicesArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServicesArgs.AsObject;
  static toObject(includeInstance: boolean, msg: GetServicesArgs): GetServicesArgs.AsObject;
  static serializeBinaryToWriter(message: GetServicesArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServicesArgs;
  static deserializeBinaryFromReader(message: GetServicesArgs, reader: jspb.BinaryReader): GetServicesArgs;
}

export namespace GetServicesArgs {
  export type AsObject = {
    serviceIdentifiersMap: Array<[string, boolean]>,
  }
}

export class GetServicesResponse extends jspb.Message {
  getServiceInfoMap(): jspb.Map<string, ServiceInfo>;
  clearServiceInfoMap(): GetServicesResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetServicesResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetServicesResponse): GetServicesResponse.AsObject;
  static serializeBinaryToWriter(message: GetServicesResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetServicesResponse;
  static deserializeBinaryFromReader(message: GetServicesResponse, reader: jspb.BinaryReader): GetServicesResponse;
}

export namespace GetServicesResponse {
  export type AsObject = {
    serviceInfoMap: Array<[string, ServiceInfo.AsObject]>,
  }
}

export class ServiceIdentifiers extends jspb.Message {
  getServiceUuid(): string;
  setServiceUuid(value: string): ServiceIdentifiers;

  getName(): string;
  setName(value: string): ServiceIdentifiers;

  getShortenedUuid(): string;
  setShortenedUuid(value: string): ServiceIdentifiers;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServiceIdentifiers.AsObject;
  static toObject(includeInstance: boolean, msg: ServiceIdentifiers): ServiceIdentifiers.AsObject;
  static serializeBinaryToWriter(message: ServiceIdentifiers, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServiceIdentifiers;
  static deserializeBinaryFromReader(message: ServiceIdentifiers, reader: jspb.BinaryReader): ServiceIdentifiers;
}

export namespace ServiceIdentifiers {
  export type AsObject = {
    serviceUuid: string,
    name: string,
    shortenedUuid: string,
  }
}

export class GetExistingAndHistoricalServiceIdentifiersResponse extends jspb.Message {
  getAllidentifiersList(): Array<ServiceIdentifiers>;
  setAllidentifiersList(value: Array<ServiceIdentifiers>): GetExistingAndHistoricalServiceIdentifiersResponse;
  clearAllidentifiersList(): GetExistingAndHistoricalServiceIdentifiersResponse;
  addAllidentifiers(value?: ServiceIdentifiers, index?: number): ServiceIdentifiers;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetExistingAndHistoricalServiceIdentifiersResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetExistingAndHistoricalServiceIdentifiersResponse): GetExistingAndHistoricalServiceIdentifiersResponse.AsObject;
  static serializeBinaryToWriter(message: GetExistingAndHistoricalServiceIdentifiersResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetExistingAndHistoricalServiceIdentifiersResponse;
  static deserializeBinaryFromReader(message: GetExistingAndHistoricalServiceIdentifiersResponse, reader: jspb.BinaryReader): GetExistingAndHistoricalServiceIdentifiersResponse;
}

export namespace GetExistingAndHistoricalServiceIdentifiersResponse {
  export type AsObject = {
    allidentifiersList: Array<ServiceIdentifiers.AsObject>,
  }
}

export class RemoveServiceArgs extends jspb.Message {
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): RemoveServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveServiceArgs): RemoveServiceArgs.AsObject;
  static serializeBinaryToWriter(message: RemoveServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveServiceArgs;
  static deserializeBinaryFromReader(message: RemoveServiceArgs, reader: jspb.BinaryReader): RemoveServiceArgs;
}

export namespace RemoveServiceArgs {
  export type AsObject = {
    serviceIdentifier: string,
  }
}

export class RemoveServiceResponse extends jspb.Message {
  getServiceUuid(): string;
  setServiceUuid(value: string): RemoveServiceResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoveServiceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RemoveServiceResponse): RemoveServiceResponse.AsObject;
  static serializeBinaryToWriter(message: RemoveServiceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoveServiceResponse;
  static deserializeBinaryFromReader(message: RemoveServiceResponse, reader: jspb.BinaryReader): RemoveServiceResponse;
}

export namespace RemoveServiceResponse {
  export type AsObject = {
    serviceUuid: string,
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
  getServiceNameSetMap(): jspb.Map<string, boolean>;
  clearServiceNameSetMap(): PartitionServices;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PartitionServices.AsObject;
  static toObject(includeInstance: boolean, msg: PartitionServices): PartitionServices.AsObject;
  static serializeBinaryToWriter(message: PartitionServices, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PartitionServices;
  static deserializeBinaryFromReader(message: PartitionServices, reader: jspb.BinaryReader): PartitionServices;
}

export namespace PartitionServices {
  export type AsObject = {
    serviceNameSetMap: Array<[string, boolean]>,
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
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): ExecCommandArgs;

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
    serviceIdentifier: string,
    commandArgsList: Array<string>,
  }
}

export class PauseServiceArgs extends jspb.Message {
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): PauseServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PauseServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: PauseServiceArgs): PauseServiceArgs.AsObject;
  static serializeBinaryToWriter(message: PauseServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PauseServiceArgs;
  static deserializeBinaryFromReader(message: PauseServiceArgs, reader: jspb.BinaryReader): PauseServiceArgs;
}

export namespace PauseServiceArgs {
  export type AsObject = {
    serviceIdentifier: string,
  }
}

export class UnpauseServiceArgs extends jspb.Message {
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): UnpauseServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnpauseServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: UnpauseServiceArgs): UnpauseServiceArgs.AsObject;
  static serializeBinaryToWriter(message: UnpauseServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnpauseServiceArgs;
  static deserializeBinaryFromReader(message: UnpauseServiceArgs, reader: jspb.BinaryReader): UnpauseServiceArgs;
}

export namespace UnpauseServiceArgs {
  export type AsObject = {
    serviceIdentifier: string,
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
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): WaitForHttpGetEndpointAvailabilityArgs;

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
    serviceIdentifier: string,
    port: number,
    path: string,
    initialDelayMilliseconds: number,
    retries: number,
    retriesDelayMilliseconds: number,
    bodyText: string,
  }
}

export class WaitForHttpPostEndpointAvailabilityArgs extends jspb.Message {
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): WaitForHttpPostEndpointAvailabilityArgs;

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
    serviceIdentifier: string,
    port: number,
    path: string,
    requestBody: string,
    initialDelayMilliseconds: number,
    retries: number,
    retriesDelayMilliseconds: number,
    bodyText: string,
  }
}

export class StreamedDataChunk extends jspb.Message {
  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): StreamedDataChunk;

  getPreviousChunkHash(): string;
  setPreviousChunkHash(value: string): StreamedDataChunk;

  getMetadata(): DataChunkMetadata | undefined;
  setMetadata(value?: DataChunkMetadata): StreamedDataChunk;
  hasMetadata(): boolean;
  clearMetadata(): StreamedDataChunk;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamedDataChunk.AsObject;
  static toObject(includeInstance: boolean, msg: StreamedDataChunk): StreamedDataChunk.AsObject;
  static serializeBinaryToWriter(message: StreamedDataChunk, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamedDataChunk;
  static deserializeBinaryFromReader(message: StreamedDataChunk, reader: jspb.BinaryReader): StreamedDataChunk;
}

export namespace StreamedDataChunk {
  export type AsObject = {
    data: Uint8Array | string,
    previousChunkHash: string,
    metadata?: DataChunkMetadata.AsObject,
  }
}

export class DataChunkMetadata extends jspb.Message {
  getName(): string;
  setName(value: string): DataChunkMetadata;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DataChunkMetadata.AsObject;
  static toObject(includeInstance: boolean, msg: DataChunkMetadata): DataChunkMetadata.AsObject;
  static serializeBinaryToWriter(message: DataChunkMetadata, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DataChunkMetadata;
  static deserializeBinaryFromReader(message: DataChunkMetadata, reader: jspb.BinaryReader): DataChunkMetadata;
}

export namespace DataChunkMetadata {
  export type AsObject = {
    name: string,
  }
}

export class UploadFilesArtifactArgs extends jspb.Message {
  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): UploadFilesArtifactArgs;

  getName(): string;
  setName(value: string): UploadFilesArtifactArgs;

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
    name: string,
  }
}

export class UploadFilesArtifactResponse extends jspb.Message {
  getUuid(): string;
  setUuid(value: string): UploadFilesArtifactResponse;

  getName(): string;
  setName(value: string): UploadFilesArtifactResponse;

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
    name: string,
  }
}

export class DownloadFilesArtifactArgs extends jspb.Message {
  getIdentifier(): string;
  setIdentifier(value: string): DownloadFilesArtifactArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadFilesArtifactArgs.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadFilesArtifactArgs): DownloadFilesArtifactArgs.AsObject;
  static serializeBinaryToWriter(message: DownloadFilesArtifactArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadFilesArtifactArgs;
  static deserializeBinaryFromReader(message: DownloadFilesArtifactArgs, reader: jspb.BinaryReader): DownloadFilesArtifactArgs;
}

export namespace DownloadFilesArtifactArgs {
  export type AsObject = {
    identifier: string,
  }
}

export class DownloadFilesArtifactResponse extends jspb.Message {
  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): DownloadFilesArtifactResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadFilesArtifactResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadFilesArtifactResponse): DownloadFilesArtifactResponse.AsObject;
  static serializeBinaryToWriter(message: DownloadFilesArtifactResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadFilesArtifactResponse;
  static deserializeBinaryFromReader(message: DownloadFilesArtifactResponse, reader: jspb.BinaryReader): DownloadFilesArtifactResponse;
}

export namespace DownloadFilesArtifactResponse {
  export type AsObject = {
    data: Uint8Array | string,
  }
}

export class StoreWebFilesArtifactArgs extends jspb.Message {
  getUrl(): string;
  setUrl(value: string): StoreWebFilesArtifactArgs;

  getName(): string;
  setName(value: string): StoreWebFilesArtifactArgs;

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
    name: string,
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
  getServiceIdentifier(): string;
  setServiceIdentifier(value: string): StoreFilesArtifactFromServiceArgs;

  getSourcePath(): string;
  setSourcePath(value: string): StoreFilesArtifactFromServiceArgs;

  getName(): string;
  setName(value: string): StoreFilesArtifactFromServiceArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreFilesArtifactFromServiceArgs.AsObject;
  static toObject(includeInstance: boolean, msg: StoreFilesArtifactFromServiceArgs): StoreFilesArtifactFromServiceArgs.AsObject;
  static serializeBinaryToWriter(message: StoreFilesArtifactFromServiceArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreFilesArtifactFromServiceArgs;
  static deserializeBinaryFromReader(message: StoreFilesArtifactFromServiceArgs, reader: jspb.BinaryReader): StoreFilesArtifactFromServiceArgs;
}

export namespace StoreFilesArtifactFromServiceArgs {
  export type AsObject = {
    serviceIdentifier: string,
    sourcePath: string,
    name: string,
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

export class RenderTemplatesToFilesArtifactArgs extends jspb.Message {
  getTemplatesAndDataByDestinationRelFilepathMap(): jspb.Map<string, RenderTemplatesToFilesArtifactArgs.TemplateAndData>;
  clearTemplatesAndDataByDestinationRelFilepathMap(): RenderTemplatesToFilesArtifactArgs;

  getName(): string;
  setName(value: string): RenderTemplatesToFilesArtifactArgs;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenderTemplatesToFilesArtifactArgs.AsObject;
  static toObject(includeInstance: boolean, msg: RenderTemplatesToFilesArtifactArgs): RenderTemplatesToFilesArtifactArgs.AsObject;
  static serializeBinaryToWriter(message: RenderTemplatesToFilesArtifactArgs, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenderTemplatesToFilesArtifactArgs;
  static deserializeBinaryFromReader(message: RenderTemplatesToFilesArtifactArgs, reader: jspb.BinaryReader): RenderTemplatesToFilesArtifactArgs;
}

export namespace RenderTemplatesToFilesArtifactArgs {
  export type AsObject = {
    templatesAndDataByDestinationRelFilepathMap: Array<[string, RenderTemplatesToFilesArtifactArgs.TemplateAndData.AsObject]>,
    name: string,
  }

  export class TemplateAndData extends jspb.Message {
    getTemplate(): string;
    setTemplate(value: string): TemplateAndData;

    getDataAsJson(): string;
    setDataAsJson(value: string): TemplateAndData;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TemplateAndData.AsObject;
    static toObject(includeInstance: boolean, msg: TemplateAndData): TemplateAndData.AsObject;
    static serializeBinaryToWriter(message: TemplateAndData, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TemplateAndData;
    static deserializeBinaryFromReader(message: TemplateAndData, reader: jspb.BinaryReader): TemplateAndData;
  }

  export namespace TemplateAndData {
    export type AsObject = {
      template: string,
      dataAsJson: string,
    }
  }

}

export class RenderTemplatesToFilesArtifactResponse extends jspb.Message {
  getUuid(): string;
  setUuid(value: string): RenderTemplatesToFilesArtifactResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenderTemplatesToFilesArtifactResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RenderTemplatesToFilesArtifactResponse): RenderTemplatesToFilesArtifactResponse.AsObject;
  static serializeBinaryToWriter(message: RenderTemplatesToFilesArtifactResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenderTemplatesToFilesArtifactResponse;
  static deserializeBinaryFromReader(message: RenderTemplatesToFilesArtifactResponse, reader: jspb.BinaryReader): RenderTemplatesToFilesArtifactResponse;
}

export namespace RenderTemplatesToFilesArtifactResponse {
  export type AsObject = {
    uuid: string,
  }
}

export class FilesArtifactNameAndUuid extends jspb.Message {
  getFilename(): string;
  setFilename(value: string): FilesArtifactNameAndUuid;

  getFileuuid(): string;
  setFileuuid(value: string): FilesArtifactNameAndUuid;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FilesArtifactNameAndUuid.AsObject;
  static toObject(includeInstance: boolean, msg: FilesArtifactNameAndUuid): FilesArtifactNameAndUuid.AsObject;
  static serializeBinaryToWriter(message: FilesArtifactNameAndUuid, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FilesArtifactNameAndUuid;
  static deserializeBinaryFromReader(message: FilesArtifactNameAndUuid, reader: jspb.BinaryReader): FilesArtifactNameAndUuid;
}

export namespace FilesArtifactNameAndUuid {
  export type AsObject = {
    filename: string,
    fileuuid: string,
  }
}

export class ListFilesArtifactNamesAndUuidsResponse extends jspb.Message {
  getFileNamesAndUuidsList(): Array<FilesArtifactNameAndUuid>;
  setFileNamesAndUuidsList(value: Array<FilesArtifactNameAndUuid>): ListFilesArtifactNamesAndUuidsResponse;
  clearFileNamesAndUuidsList(): ListFilesArtifactNamesAndUuidsResponse;
  addFileNamesAndUuids(value?: FilesArtifactNameAndUuid, index?: number): FilesArtifactNameAndUuid;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListFilesArtifactNamesAndUuidsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListFilesArtifactNamesAndUuidsResponse): ListFilesArtifactNamesAndUuidsResponse.AsObject;
  static serializeBinaryToWriter(message: ListFilesArtifactNamesAndUuidsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListFilesArtifactNamesAndUuidsResponse;
  static deserializeBinaryFromReader(message: ListFilesArtifactNamesAndUuidsResponse, reader: jspb.BinaryReader): ListFilesArtifactNamesAndUuidsResponse;
}

export namespace ListFilesArtifactNamesAndUuidsResponse {
  export type AsObject = {
    fileNamesAndUuidsList: Array<FilesArtifactNameAndUuid.AsObject>,
  }
}

