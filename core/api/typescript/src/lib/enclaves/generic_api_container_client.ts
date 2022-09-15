/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import { Result } from "neverthrow";
import {
    ExecCommandArgs,
    ExecCommandResponse,
    ExecuteModuleArgs,
    ExecuteModuleResponse,
    GetModulesArgs,
    GetModulesResponse,
    GetServicesArgs,
    GetServicesResponse,
    LoadModuleArgs,
    PauseServiceArgs,
    RemoveServiceArgs,
    RemoveServiceResponse,
    RenderTemplatesToFilesArtifactArgs,
    RenderTemplatesToFilesArtifactResponse,
    RepartitionArgs,
    StartServicesArgs,
    StartServicesResponse,
    StoreFilesArtifactFromServiceArgs,
    StoreWebFilesArtifactArgs,
    StoreWebFilesArtifactResponse,
    UnloadModuleArgs,
    UnloadModuleResponse,
    UnpauseServiceArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { EnclaveID } from "./enclave_context";

export interface GenericApiContainerClient {
    getEnclaveId(): EnclaveID
    loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<null, Error>>
    unloadModule(unloadModuleArgs: UnloadModuleArgs): Promise<Result<UnloadModuleResponse,Error>>
    startServices(startServicesArgs: StartServicesArgs): Promise<Result<StartServicesResponse, Error>>
    removeService(args: RemoveServiceArgs): Promise<Result<RemoveServiceResponse, Error>>
    repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>>
    waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>>
    getServices(getServicesArgs: GetServicesArgs): Promise<Result<GetServicesResponse, Error>>
    getModules(getModulesArgs: GetModulesArgs): Promise<Result<GetModulesResponse, Error>>
    executeModule(executeModuleArgs: ExecuteModuleArgs): Promise<Result<ExecuteModuleResponse, Error>>
    execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>>
    pauseService(pauseServiceArgs: PauseServiceArgs): Promise<Result<null, Error>>
    unpauseService(unpauseServiceArgs: UnpauseServiceArgs): Promise<Result<null, Error>>
    uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>>
    storeWebFilesArtifact(storeWebFilesArtifactArgs: StoreWebFilesArtifactArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
    storeFilesArtifactFromService(storeFilesArtifactFromServiceArgs: StoreFilesArtifactFromServiceArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
    renderTemplatesToFilesArtifact(renderTemplatesToFilesArtifactArgs: RenderTemplatesToFilesArtifactArgs): Promise<Result<RenderTemplatesToFilesArtifactResponse, Error>>
}