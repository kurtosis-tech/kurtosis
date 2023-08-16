//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {GetServiceLogsArgs, CreateEnclaveArgs} from 'kurtosis-sdk/build/engine/kurtosis_engine_rpc_api_bindings/engine_service_pb'
import {runStarlarkPackage} from "./container"
import axios from "axios";

const ENGINE_URL =  "http://localhost:9710"

const createApiPromiseClient = (apiClient) => {
    if (apiClient) {
        return `http://localhost:${apiClient.grpcPortOnHostMachine}`
    }
    return "";
}

export const makeRestApiRequest = async ( url, data, config) => {
    const response = await axios.post(`${ENGINE_URL}/${url}`, data, config)
    return response;
}

export const getEnclavesFromKurtosis = async () => {
    const respFromGrpc = await makeRestApiRequest(
         "engine_api.EngineService/GetEnclaves",
        {"field":""},
        {"headers":{'Content-Type': "application/json"}}
    )

    const {data} = respFromGrpc
    return Object.keys(data.enclaveInfo).map(key => {
        const enclave = data.enclaveInfo[key]
        return {
            uuid: enclave.enclaveUuid,
            name: enclave.name,
            created: enclave.creationTime,
            status: enclave.apiContainerStatus,
            apiClient: createApiPromiseClient(enclave.apiContainerHostMachineInfo)
        }
    })

}

export const createEnclave = async () => {
    const enclaveArgs = new CreateEnclaveArgs();
    enclaveArgs.setApiContainerVersionTag("")
    enclaveArgs.setApiContainerLogLevel("info");
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



export const runStarlark = async(apiClient, packageId, args) => {
    const stream = await runStarlarkPackage(apiClient, packageId, args)
    return stream;
}
