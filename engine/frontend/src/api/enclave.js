//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {GetServiceLogsArgs, CreateEnclaveArgs} from 'kurtosis-sdk/build/engine/kurtosis_engine_rpc_api_bindings/engine_service_pb'
import {EngineServicePromiseClient} from "kurtosis-sdk/build/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb"
//import {ApiContainerServiceClient} from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb'
import google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb.js'
import { StoreFilesArtifactFromServiceArgs } from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_pb';
import {runStarlarkPackage} from "./container"

const engineClient = new EngineServicePromiseClient("http://localhost:9710");

const createApiPromiseClient = (apiClient) => {
    if (apiClient) {
        return `http://localhost:${apiClient.grpcPortOnHostMachine}`
    }
    return "";
}

export const getEnclavesFromKurtosis = async () => {
    const respFromGrpc = await engineClient.getEnclaves(new google_protobuf_empty_pb.Empty, null);
    const response = respFromGrpc.toObject()
    
    // processing the data so that frontend can consume it! 
    const responseProcessed = response.enclaveInfoMap.map(enclave => {
        return {
            uuid: enclave[0],
            name: enclave[1].name,
            created: enclave[1].creationTime.seconds,
            status: enclave[1].apiContainerStatus,
            apiClient: createApiPromiseClient(enclave[1].apiContainerHostMachineInfo) 
        }
    });

    return responseProcessed
}

export const getServiceLogs = async (enclaveName, serviceUuid) => {
    const args = new GetServiceLogsArgs();

    args.setEnclaveIdentifier(enclaveName);
    const serviceUuidMapSet = args.getServiceUuidSetMap();
    const isServiceUuidInSet = true;
    serviceUuidMapSet.set(serviceUuid, isServiceUuidInSet)
    args.setFollowLogs(true);

    const stream = engineClient.getServiceLogs(args, {});
    return stream;
}

export const createEnclave = async () => {
    const enclaveArgs = new CreateEnclaveArgs();
    enclaveArgs.setApiContainerVersionTag("")
    enclaveArgs.setApiContainerLogLevel("info");
    enclaveArgs.setIsPartitioningEnabled(false);
    const enclaveGRPC = await engineClient.createEnclave(enclaveArgs, null)
    const enclave = enclaveGRPC.toObject().enclaveInfo;
    return {
        uuid: enclave.uuid,
        name: enclave.name,
        created: enclave.creationTime.seconds,
        status: enclave.apiContainerStatus,
        apiClient: createApiPromiseClient(enclave.apiContainerHostMachineInfo) 
    }
}

export const runStarlark = async(apiClient, packageId, args) => {
    const stream = await runStarlarkPackage(apiClient, packageId, args)
    return stream;
}
