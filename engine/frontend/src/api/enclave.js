//import {EngineServicePromiseClient} from 'kurtosis-sdk/src/engine/kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb'
import {runStarlarkPackage} from "./container"
import axios from "axios";

import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";

import {createPromiseClient} from "@bufbuild/connect";

import {createConnectTransport,} from "@bufbuild/connect-web";
import {createEnclaveFromEnclaveManager, getEnclavesFromEnclaveManager} from "./api";

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

export const getEnclavesFromKurtosis = async (token) => {
    const data = await getEnclavesFromEnclaveManager(token);
    if ("enclaveInfo" in data) {
        return Object.keys(data.enclaveInfo).map(key => {
            const enclave = data.enclaveInfo[key]
            return {
                uuid: enclave.enclaveUuid,
                name: enclave.name,
                // created: enclave.creationTime,
                status: enclave.apiContainerStatus,
                host: enclave.apiContainerInfo.bridgeIpAddress,
                port: enclave.apiContainerInfo.grpcPortInsideEnclave,
            }
        });
    }
    return []
}

export const createEnclave = async (token) => {
    const enclaveName = ""; // TODO We could make this input from the UI
    const apiContainerVersionTag = "";
    const apiContainerLogLevel = "info";
    const response = await createEnclaveFromEnclaveManager(enclaveName, apiContainerLogLevel, apiContainerVersionTag, token)

    const enclave = response.enclaveInfo;
    return {
        uuid: enclave.enclaveUuid,
        name: enclave.name,
        created: enclave.creationTime,
        status: enclave.apiContainerStatus,
        host: enclave.apiContainerInfo.bridgeIpAddress,
        port: enclave.apiContainerInfo.grpcPortInsideEnclave,
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

export const runStarlark = async (host, port, packageId, args, token) => {
    const stream = await runStarlarkPackage(host, port, packageId, args, token)
    return stream;
}
