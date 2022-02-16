import { Result } from "neverthrow";
import { ApiContainerServiceClientNode, ApiContainerServiceClientWeb } from "kurtosis-core-api-lib";
import { 
    CleanArgs, 
    CleanResponse, 
    CreateEnclaveArgs, 
    CreateEnclaveResponse, 
    DestroyEnclaveArgs, 
    EnclaveAPIContainerHostMachineInfo, 
    GetEnclavesResponse, 
    GetEngineInfoResponse, 
    StopEnclaveArgs 
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";

export interface KurtosisContextBackend {
    getEngineInfo(): Promise<Result<GetEngineInfoResponse,Error>>
    createEnclaveResponse(args: CreateEnclaveArgs): Promise<Result<CreateEnclaveResponse, Error>>
    getEnclavesResponse(): Promise<Result<GetEnclavesResponse, Error>>
    stopEnclave(stopEnclaveArgs: StopEnclaveArgs): Promise<Result<null, Error>>
    destroyEnclave(destroyEnclaveArgs: DestroyEnclaveArgs): Promise<Result<null, Error>>
    clean(cleanArgs: CleanArgs): Promise<Result<CleanResponse, Error>>
    createApiClient(localhostIpAddress:string, apiContainerHostMachineInfo:EnclaveAPIContainerHostMachineInfo):Result<ApiContainerServiceClientWeb | ApiContainerServiceClientNode,Error>
}