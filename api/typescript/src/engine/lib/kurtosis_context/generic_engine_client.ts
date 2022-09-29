import { Result } from "neverthrow";
import {
    CleanArgs,
    CleanResponse,
    CreateEnclaveArgs,
    CreateEnclaveResponse,
    DestroyEnclaveArgs,
    GetEnclavesResponse,
    GetEngineInfoResponse,
    StopEnclaveArgs
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";

export interface GenericEngineClient {
    getEngineInfo(): Promise<Result<GetEngineInfoResponse,Error>>
    createEnclaveResponse(args: CreateEnclaveArgs): Promise<Result<CreateEnclaveResponse, Error>>
    getEnclavesResponse(): Promise<Result<GetEnclavesResponse, Error>>
    stopEnclave(stopEnclaveArgs: StopEnclaveArgs): Promise<Result<null, Error>>
    destroyEnclave(destroyEnclaveArgs: DestroyEnclaveArgs): Promise<Result<null, Error>>
    clean(cleanArgs: CleanArgs): Promise<Result<CleanResponse, Error>>
}