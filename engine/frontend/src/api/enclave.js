//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {
    CreateEnclaveArgs,
    GetServiceLogsArgs
} from 'kurtosis-sdk/build/engine/kurtosis_engine_rpc_api_bindings/engine_service_pb'
import {
    EngineServicePromiseClient
} from "kurtosis-sdk/build/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb"
//import {ApiContainerServiceClient} from 'kurtosis-sdk/build/core/kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb'
import {runStarlarkPackage} from "./container"
import axios from "axios"


const engineClient = new EngineServicePromiseClient("http://localhost:9710");

const createApiPromiseClient = (apiClient) => {
    if (apiClient) {
        return `http://localhost:${apiClient.grpcPortOnHostMachine}`
    }
    return "";
}


export const getEnclavesFromKurtosis = async () => {
    const response = await axios.post("http://localhost:9710/engine_api.EngineService/GetEnclaves", {"field":""}, {"headers":{'Content-Type': "application/json"}})
    console.log(typeof (response.data.enclaveInfo))
    // const respFromGrpc = await engineClient.getEnclaves(new google_protobuf_empty_pb.Empty, null);
    // const response = respFromGrpc.toObject()
    
    // processing the data so that frontend can consume it! 

    return Object.keys(response.data.enclaveInfo).map(key => {
        console.log(key)
        const enclave = response.data.enclaveInfo[key]
        console.log(enclave)
        return {
            uuid: enclave.enclaveUuid,
            name: enclave.name,
            created: enclave.creationTime,
            status: enclave.apiContainerStatus,
            apiClient: createApiPromiseClient(enclave.apiContainerHostMachineInfo)
        }
    })
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
    const apiClient = createApiPromiseClient(enclave.apiContainerHostMachineInfo);
    return {enclave, apiClient}
}

export const runStarlark = async(apiClient, packageId) => {
    const stream = await runStarlarkPackage(apiClient, packageId)
    return stream;
}
