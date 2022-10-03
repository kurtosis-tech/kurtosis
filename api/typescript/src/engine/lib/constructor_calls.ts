import {
    CleanArgs,
    CreateEnclaveArgs,
    DestroyEnclaveArgs,
    GetUserServiceLogsArgs,
    StopEnclaveArgs
} from "../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import * as jspb from "google-protobuf";
import {ServiceGUID} from "../../core/lib/services/service";

// ====================================================================================================
//                                    Kurtosis Context
// ====================================================================================================


export function newCreateEnclaveArgs(
        enclaveId: string,
        apiContainerImageVersionTag: string,
        apiContainerLogLevel: string,
        isPartitioningEnabled: boolean): CreateEnclaveArgs {
    const result: CreateEnclaveArgs = new CreateEnclaveArgs();
    result.setEnclaveId(enclaveId);
    result.setApiContainerVersionTag(apiContainerImageVersionTag);
    result.setApiContainerLogLevel(apiContainerLogLevel);
    result.setIsPartitioningEnabled(isPartitioningEnabled);

    return result;
}

export function newStopEnclaveArgs(enclaveId:string): DestroyEnclaveArgs {
    const result: StopEnclaveArgs = new StopEnclaveArgs();
    result.setEnclaveId(enclaveId);
    return result;
}

export function newDestroyEnclaveArgs(enclaveId:string): DestroyEnclaveArgs {
    const result: DestroyEnclaveArgs = new DestroyEnclaveArgs();
    result.setEnclaveId(enclaveId);
    return result;
}

export function newCleanArgs(shouldCleanAll:boolean): CleanArgs {
    const result: CleanArgs = new CleanArgs();
    result.setShouldCleanAll(shouldCleanAll);
    return result;
}

export function newGetUserServiceLogsArgs(
        enclaveId: string,
        userServiceGuids: Set<ServiceGUID>): GetUserServiceLogsArgs {

    const result: GetUserServiceLogsArgs = new GetUserServiceLogsArgs();
    result.setEnclaveId(enclaveId)
    const serviceGuidSetMap: jspb.Map<string, boolean> = result.getServiceGuidSetMap();
    for (const serviceGuid of userServiceGuids) {
        serviceGuidSetMap.set(serviceGuid, true);
    }
    return result;
}
