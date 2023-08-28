import {runStarlarkPackage} from "./container"
import axios from "axios";

import {createClient, createEnclaveFromEnclaveManager, getEnclavesFromEnclaveManager} from "./api";

export const getKurtosisPackages = async () => {
    const response = await axios.post(`http://localhost:9770/kurtosis_package_indexer.KurtosisPackageIndexer/GetPackages`, {"field":""}, {"headers":{'Content-Type': "application/json"}}
    )
    const {data} = response
    return data.packages
}

export const getEnclavesFromKurtosis = async (token, apiHost) => {
    const data = await getEnclavesFromEnclaveManager(token, apiHost);
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

export const createEnclave = async (token, apiHost) => {
    const enclaveName = ""; // TODO We could make this input from the UI
    const apiContainerVersionTag = "";
    const apiContainerLogLevel = "info";
    const response = await createEnclaveFromEnclaveManager(enclaveName, apiContainerLogLevel, apiContainerVersionTag, token, apiHost)

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

export const getServiceLogs = async (ctrl, enclaveName, serviceUuid, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const args = {
        "enclaveIdentifier": enclaveName,
        "serviceUuidSet": {
            [serviceUuid]: true
        },
        followLogs: true,
    }
    return enclaveManagerClient.getServiceLogs(args, {signal: ctrl.signal});
}

export const runStarlark = async (host, port, packageId, args, token, apiHost) => {
    const stream = await runStarlarkPackage(host, port, packageId, args, token, apiHost)
    return stream;
}
