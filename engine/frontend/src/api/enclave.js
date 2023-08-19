//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {runStarlarkPackage} from "./container"
import axios from "axios";

import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";

import {createPromiseClient} from "@bufbuild/connect";

import {createConnectTransport,} from "@bufbuild/connect-web";
import {getEnclavesFromEnclaveManager} from "./api";


const transport = createConnectTransport({
    baseUrl: "http://localhost:8081"
})

const enclaveManagerClient = createPromiseClient(KurtosisEnclaveManagerServer, transport);
const ENGINE_URL = "http://localhost:8081"

const createApiPromiseClient = (apiClient) => {
    if (apiClient) {
        return `http://localhost:${apiClient.grpcPortOnHostMachine}`
    }
    return "";
}

export const makeRestApiRequest = async (url, data, config) => {
    const response = await axios.post(`${ENGINE_URL}/${url}`, data, config)
    return response;
}

export const getEnclavesFromKurtosis = async () => {
    const data = await getEnclavesFromEnclaveManager();
    if ("enclaveInfo" in data) {
        return Object.keys(data.enclaveInfo).map(key => {
            const enclave = data.enclaveInfo[key]
            console.log("enclave",enclave)
            return {
                uuid: enclave.enclaveUuid,
                name: enclave.name,
                // created: enclave.creationTime,
                status: enclave.apiContainerStatus,
                host: enclave.apiContainerHostMachineInfo.ipOnHostMachine,
                port: enclave.apiContainerHostMachineInfo.grpcPortOnHostMachine,
            }
        });
    }
    return []
}

export const createEnclave = async () => {
    const data = {
        apiContainerVersionTag: "",
        apiContainerLogLevel: "info",
        isPartitioningEnabled: false,
    }
    const response = await makeRestApiRequest("engine_api.EngineService/CreateEnclave", JSON.stringify(data), {"headers": {'Content-Type': "application/json"}})

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

export const getServiceLogs = async (ctrl, enclaveName, serviceUuid) => {
    const args = {
        "enclaveIdentifier": enclaveName,
        "serviceUuidSet": {
            [serviceUuid]: true
        },
        followLogs: true,
    }
    return enclaveManagerClient.getServiceLogs(args, {signal: ctrl.signal});
}

export const runStarlark = async (apiClient, packageId, args) => {
    const stream = await runStarlarkPackage(apiClient, packageId, args)
    return stream;
}
