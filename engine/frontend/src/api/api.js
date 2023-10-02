import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {
    GetListFilesArtifactNamesAndUuidsRequest,
    GetServicesRequest,
    InspectFilesArtifactContentsRequest,
    RunStarlarkPackageRequest
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import {CreateEnclaveArgs, DestroyEnclaveArgs, EnclaveMode} from "enclave-manager-sdk/build/engine_service_pb";
import {RunStarlarkPackageArgs} from "enclave-manager-sdk/build/api_container_service_pb";

export const createClient = (apiHost) => {
    if (apiHost && apiHost.length > 0) {
        const transport = createConnectTransport({baseUrl: apiHost})
        return createPromiseClient(KurtosisEnclaveManagerServer, transport)
    }
    throw "no ApiHost provided"
}

const createHeaderOptionsWithToken = (token) => {
    const headers = new Headers();
    if (token && token.length > 0) {
        headers.set("Authorization", `Bearer ${token}`);
        return {headers: headers}
    }
    return {}
}

export const getEnclavesFromEnclaveManager = async (token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    return enclaveManagerClient.getEnclaves({}, createHeaderOptionsWithToken(token));
}

export const getServicesFromEnclaveManager = async (host, port, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new GetServicesRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.getServices(request, createHeaderOptionsWithToken(token));
}

export const listFilesArtifactNamesAndUuidsFromEnclaveManager = async (host, port, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new GetListFilesArtifactNamesAndUuidsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.listFilesArtifactNamesAndUuids(request, createHeaderOptionsWithToken(token));
}

export const inspectFilesArtifactContentsFromEnclaveManager = async (host, port, fileName, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
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

export const removeEnclaveFromEnclaveManager = async (enclaveIdentifier, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new DestroyEnclaveArgs(
        {
            "enclaveIdentifier": enclaveIdentifier
        }
    );
    return enclaveManagerClient.destroyEnclave(request, createHeaderOptionsWithToken(token))
}

export const createEnclaveFromEnclaveManager = async (enclaveName, logLevel, versionTag, token, apiHost, productionMode) => {
    const enclaveManagerClient = createClient(apiHost);
    const mode = productionMode ? EnclaveMode.PRODUCTION : EnclaveMode.TEST; 
    const request = new CreateEnclaveArgs(
        {
            "enclaveName": enclaveName,
            "apiContainerVersionTag": versionTag,
            "apiContainerLogLevel": logLevel,
            "mode": mode,
        }
    );
    console.log("Sending Create Enclave Request with Args", request)
    return enclaveManagerClient.createEnclave(request, createHeaderOptionsWithToken(token));
}

export const runStarlarkPackageFromEnclaveManager = async (host, port, packageId, args, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    
    const runStarlarkPackageArgs = new RunStarlarkPackageArgs(
        {
            "dryRun": false,
            "remote": "RunStarlarkPackageArgs_Remote",
            "packageId": packageId,
            "serializedParams": args,
        }
    )

    const request = new RunStarlarkPackageRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
            "RunStarlarkPackageArgs": runStarlarkPackageArgs
        }
    );
    return enclaveManagerClient.runStarlarkPackage(request, createHeaderOptionsWithToken(token));
}


export const getStarlarkRunConfig = async (host, port, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new InspectFilesArtifactContentsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.inspectFilesArtifactContents(request, createHeaderOptionsWithToken(token));
}
