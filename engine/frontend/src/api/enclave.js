import {
    createClient,
    createEnclaveFromEnclaveManager,
    getEnclavesFromEnclaveManager,
    removeEnclaveFromEnclaveManager,
    runStarlarkPackageFromEnclaveManager
} from "./api";

export const getEnclavesFromKurtosis = async (token, apiHost) => {
    const data = await getEnclavesFromEnclaveManager(token, apiHost);
    if ("enclaveInfo" in data) {
        return Object.keys(data.enclaveInfo).map(key => {
            const enclave = data.enclaveInfo[key]
            if (!enclave.containersStatus) {
                return {}
            }
            return {
                uuid: enclave.enclaveUuid,
                name: enclave.name,
                created: enclave.creationTime,
                status: enclave.apiContainerStatus,
                host: enclave.apiContainerInfo.bridgeIpAddress,
                port: enclave.apiContainerInfo.grpcPortInsideEnclave,
                mode: enclave.mode,
            }
        });
    }
    return []
}

export const removeEnclave = async (token, apiHost, enclaveName) => {
    const response = await removeEnclaveFromEnclaveManager(enclaveName, token, apiHost)
    const enclave = response.enclaveInfo;
    return {}
}


export const createEnclave = async (token, apiHost, enclaveName, productionMode) => {
    const apiContainerVersionTag = "";
    const apiContainerLogLevel = "info";
    const response = await createEnclaveFromEnclaveManager(enclaveName, apiContainerLogLevel, apiContainerVersionTag, token, apiHost, productionMode)

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

export const getServiceLogs = async (ctrl, enclaveName, serviceUuid, apiHost, followLogs, numLogLines) => {
    const enclaveManagerClient = createClient(apiHost);
    const args = {
        enclaveIdentifier: enclaveName,
        serviceUuidSet: {
            [serviceUuid]: true
        },
        followLogs: followLogs,
        returnAllLogs: false,
        numLogLines: numLogLines,
    }
    console.log("Sending GetServiceLogs Request with Args", args)
    return enclaveManagerClient.getServiceLogs(args, {signal: ctrl.signal});
}

export const runStarlark = async (host, port, packageId, args, token, apiHost) => {
    const stream = await runStarlarkPackageFromEnclaveManager(host, port, packageId, args, token, apiHost)
    return stream;
}
