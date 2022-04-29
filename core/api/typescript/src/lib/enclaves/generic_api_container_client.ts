/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import { Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {
    ExecCommandArgs,
    ExecCommandResponse,
    ExecuteModuleArgs,
    ExecuteModuleResponse,
    GetModuleInfoArgs,
    GetModulesResponse,
    GetServiceInfoArgs,
    GetServiceInfoResponse,
    GetServicesResponse,
    LoadModuleArgs,
    RegisterServiceArgs,
    RegisterServiceResponse,
    RemoveServiceArgs,
    RepartitionArgs,
    StartServiceArgs,
    StartServiceResponse, StoreFilesArtifactFromServiceArgs, StoreWebFilesArtifactArgs, StoreWebFilesArtifactResponse,
    UnloadModuleArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { EnclaveID } from "./enclave_context";

export interface GenericApiContainerClient {
    getEnclaveId(): EnclaveID
    loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<null, Error>>
    unloadModule(unloadModuleArgs: UnloadModuleArgs): Promise<Result<null,Error>>
    getModuleInfo(getModuleInfoArgs: GetModuleInfoArgs): Promise<Result<null, Error>>
    registerService(registerServiceArgs: RegisterServiceArgs): Promise<Result<RegisterServiceResponse, Error>>
    startService(startServiceArgs: StartServiceArgs): Promise<Result<StartServiceResponse, Error>>
    getServiceInfo(getServiceInfoArgs: GetServiceInfoArgs): Promise<Result<GetServiceInfoResponse, Error>>
    removeService(args: RemoveServiceArgs): Promise<Result<null, Error>>
    repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>>
    waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>>
    getServices(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetServicesResponse, Error>>
    getModules(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetModulesResponse, Error>>
    executeModule(executeModuleArgs: ExecuteModuleArgs): Promise<Result<ExecuteModuleResponse, Error>>
    execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>>
    uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>>
    storeWebFilesArtifact(storeWebFilesArtifactArgs: StoreWebFilesArtifactArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
    storeFilesArtifactFromService(storeFilesArtifactFromServiceArgs: StoreFilesArtifactFromServiceArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
}