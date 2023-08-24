import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {
    GetListFilesArtifactNamesAndUuidsRequest,
    GetServicesRequest,
    InspectFilesArtifactContentsRequest
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import {CreateEnclaveArgs} from "enclave-manager-sdk/build/engine_service_pb";
import {RunStarlarkPackageArgs} from "enclave-manager-sdk/build/api_container_service_pb";

export const createClient = (apiHost) => {
    const transport = createConnectTransport({baseUrl: apiHost})
    return createPromiseClient(KurtosisEnclaveManagerServer, transport)
}

const createHeaderOptionsWithToken = (token) => {
    const headers = new Headers();
    if (token) {
        headers.set("Authorization", `Bearer ${token}`);
    }
    return {headers: headers}
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


export const createEnclaveFromEnclaveManager = async (enclaveName, logLevel, versionTag, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new CreateEnclaveArgs(
        {
            "enclaveName": enclaveName,
            "apiContainerVersionTag": versionTag,
            "apiContainerLogLevel": logLevel,
        }
    );
    return enclaveManagerClient.createEnclave(request, createHeaderOptionsWithToken(token));
}

export const runStarlarkPackageFromEnclaveManager = async (host, port, packageId, args, token, apiHost) => {
    const enclaveManagerClient = createClient(apiHost);
    const request = new RunStarlarkPackageArgs(
        {
            "dryRun": false,
            "remote": "RunStarlarkPackageArgs_Remote",
            "packageId": packageId,
            "serializedParams": args,
        }
    )
    return enclaveManagerClient.runStarlarkPackage(request, createHeaderOptionsWithToken(token));
}
