/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import { Result } from "neverthrow";
import {
    DownloadFilesArtifactArgs, DownloadFilesArtifactResponse,
    ExecCommandArgs,
    ExecCommandResponse,
    GetExistingAndHistoricalServiceIdentifiersResponse,
    GetServicesArgs,
    GetServicesResponse,
    ListFilesArtifactNamesAndUuidsResponse,
    RunStarlarkPackageArgs,
    RunStarlarkScriptArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { EnclaveUUID } from "./enclave_context";
import {Readable} from "stream";

export interface GenericApiContainerClient {
    getEnclaveUuid(): EnclaveUUID
    getEnclaveName(): string
    runStarlarkScript(serializedStarlarkScript: RunStarlarkScriptArgs): Promise<Result<Readable, Error>>
    runStarlarkPackage(starlarkPackageArgs: RunStarlarkPackageArgs): Promise<Result<Readable, Error>>
    waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>>
    getServices(getServicesArgs: GetServicesArgs): Promise<Result<GetServicesResponse, Error>>
    execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>>
    uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>>
    downloadFilesArtifact(downloadFilesArtifactArgs: DownloadFilesArtifactArgs): Promise<Result<DownloadFilesArtifactResponse, Error>>
    getExistingAndHistoricalServiceIdentifiers(): Promise<Result<GetExistingAndHistoricalServiceIdentifiersResponse, Error>>
    getAllFilesArtifactNamesAndUuids(): Promise<Result<ListFilesArtifactNamesAndUuidsResponse, Error>>
}
