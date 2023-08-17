//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {runStarlarkPackage} from "./container"
import axios from "axios";

import {EngineService} from  "kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/connect/engine_service_connect";

import {createPromiseClient} from "@bufbuild/connect";

import {
    createConnectTransport,
} from "@bufbuild/connect-web";


const transport = createConnectTransport({
    baseUrl: "http://localhost:9710"
})

const engineClient = createPromiseClient(EngineService, transport);

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
    const data = {
        apiContainerVersionTag: "",
        apiContainerLogLevel: "info",
        isPartitioningEnabled: false,
    }
    const response = await makeRestApiRequest("engine_api.EngineService/CreateEnclave", JSON.stringify(data), {"headers":{'Content-Type': "application/json"}})

    const enclave = response.data.enclaveInfo;
    const apiClient = createApiPromiseClient(enclave.apiContainerHostMachineInfo);

    return {
        uuid: enclave.enclaveUuid,
        name: enclave.name,
        created: enclave.creationTime,
        status: enclave.apiContainerStatus,
        apiClient
    }
}

export const getServiceLogs = async (enclaveName, serviceUuid) => {
    const args = {
        "enclaveIdentifier": enclaveName,
        "serviceUuidSet": {
            [serviceUuid]: true
        },
        followLogs: false,
    }
    return engineClient.getServiceLogs(args);
}

export const runStarlark = async(apiClient, packageId, args) => {
    const stream = await runStarlarkPackage(apiClient, packageId, args)
    return stream;
}
