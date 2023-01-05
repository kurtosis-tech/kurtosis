/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import { Result } from "neverthrow";
import {
    ExecCommandArgs,
    ExecCommandResponse,
    GetServicesArgs,
    GetServicesResponse,
    PauseServiceArgs,
    RemoveServiceArgs,
    RemoveServiceResponse,
    RenderTemplatesToFilesArtifactArgs,
    RenderTemplatesToFilesArtifactResponse,
    RepartitionArgs,
    RunStarlarkPackageArgs,
    RunStarlarkScriptArgs,
    StartServicesArgs,
    StartServicesResponse,
    StoreFilesArtifactFromServiceArgs,
    StoreWebFilesArtifactArgs,
    StoreWebFilesArtifactResponse,
    UnpauseServiceArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { EnclaveID } from "./enclave_context";
import {Readable} from "stream";

export interface GenericApiContainerClient {
    getEnclaveId(): EnclaveID
    runStarlarkScript(serializedStarlarkScript: RunStarlarkScriptArgs): Promise<Result<Readable, Error>>
    runStarlarkPackage(starlarkPackageArgs: RunStarlarkPackageArgs): Promise<Result<Readable, Error>>
    startServices(startServicesArgs: StartServicesArgs): Promise<Result<StartServicesResponse, Error>>
    removeService(args: RemoveServiceArgs): Promise<Result<RemoveServiceResponse, Error>>
    repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>>
    waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>>
    getServices(getServicesArgs: GetServicesArgs): Promise<Result<GetServicesResponse, Error>>
    execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>>
    pauseService(pauseServiceArgs: PauseServiceArgs): Promise<Result<null, Error>>
    unpauseService(unpauseServiceArgs: UnpauseServiceArgs): Promise<Result<null, Error>>
    uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>>
    storeWebFilesArtifact(storeWebFilesArtifactArgs: StoreWebFilesArtifactArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
    storeFilesArtifactFromService(storeFilesArtifactFromServiceArgs: StoreFilesArtifactFromServiceArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>>
    renderTemplatesToFilesArtifact(renderTemplatesToFilesArtifactArgs: RenderTemplatesToFilesArtifactArgs): Promise<Result<RenderTemplatesToFilesArtifactResponse, Error>>
}