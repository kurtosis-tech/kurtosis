import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {
    GetListFilesArtifactNamesAndUuidsRequest,
    GetServicesRequest,
    InspectFilesArtifactContentsRequest,
    RunStarlarkPackageRequest
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import {CreateEnclaveArgs} from "enclave-manager-sdk/build/engine_service_pb";

const transport = createConnectTransport({
    baseUrl: "http://localhost:8081"
})

const enclaveManagerClient = createPromiseClient(KurtosisEnclaveManagerServer, transport);

const createHeaderOptionsWithToken = (token) => {
    const headers = new Headers();
    if (token) {
        headers.set("Authorization", `Bearer ${token}`);
    }
    console.log("headers: ", headers.get("Authorization"))
    return {headers: headers}
}

export const getEnclavesFromEnclaveManager = async (token) => {
    return enclaveManagerClient.getEnclaves({}, createHeaderOptionsWithToken(token));
}

export const getServicesFromEnclaveManager = async (host, port, token) => {
    const request = new GetServicesRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.getServices(request, createHeaderOptionsWithToken(token));
}

export const listFilesArtifactNamesAndUuidsFromEnclaveManager = async (host, port, token) => {
    const request = new GetListFilesArtifactNamesAndUuidsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.listFilesArtifactNamesAndUuids(request, createHeaderOptionsWithToken(token));
}

export const inspectFilesArtifactContentsFromEnclaveManager = async (host, port, fileName, token) => {
    const request = new InspectFilesArtifactContentsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
            "fileNamesAndUuid": {
                "fileName": fileName
            }
        }
    );
    return enclaveManagerClient.inspectFilesArtifactContents(request, createHeaderOptionsWithToken(token));
}


export const createEnclaveFromEnclaveManager = async (enclaveName, logLevel, versionTag, token) => {
    const request = new CreateEnclaveArgs(
        {
            "enclaveName": enclaveName,
            "apiContainerVersionTag": versionTag,
            "apiContainerLogLevel": logLevel,
        }
    );
    return enclaveManagerClient.createEnclave(request, createHeaderOptionsWithToken(token));
}

export const runStarlarkPackageFromEnclaveManager = async (host, port, enclaveName, logLevel, versionTag, token) => {
    const request = new RunStarlarkPackageRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
            "RunStarlarkPackageArgs": {
                // TODO
            }
        }
    );
    return enclaveManagerClient.runStarlarkPackage(request, createHeaderOptionsWithToken(token));
}
