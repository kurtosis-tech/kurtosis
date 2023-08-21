import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {
    GetListFilesArtifactNamesAndUuidsRequest,
    GetServicesRequest,
    InspectFilesArtifactContentsRequest, RunStarlarkPackageRequest
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import {CreateEnclaveArgs} from "enclave-manager-sdk/build/engine_service_pb";
import {RunStarlarkPackageArgs} from "enclave-manager-sdk/build/api_container_service_pb";

const transport = createConnectTransport({
    baseUrl: "http://localhost:8081"
})

const enclaveManagerClient = createPromiseClient(KurtosisEnclaveManagerServer, transport);

export const getEnclavesFromEnclaveManager = async () => {
    return enclaveManagerClient.getEnclaves({});
}

export const getServicesFromEnclaveManager = async (host, port) => {
    const request = new GetServicesRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.getServices(request);
}

export const listFilesArtifactNamesAndUuidsFromEnclaveManager = async (host, port) => {
    const request = new GetListFilesArtifactNamesAndUuidsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    return enclaveManagerClient.listFilesArtifactNamesAndUuids(request);
}

export const inspectFilesArtifactContentsFromEnclaveManager = async (host, port, fileName, fileUuid) => {
    const request = new InspectFilesArtifactContentsRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
            "fileNamesAndUuid": {
                "fileName": fileName
            }
        }
    );
    return enclaveManagerClient.inspectFilesArtifactContents(request);
}


export const createEnclaveFromEnclaveManager = async (enclaveName, logLevel, versionTag) => {
    const request = new CreateEnclaveArgs(
        {
            "enclaveName": enclaveName,
            "apiContainerVersionTag": versionTag,
            "apiContainerLogLevel": logLevel,
        }
    );
    return enclaveManagerClient.createEnclave(request);
}

export const runStarlarkPackageFromEnclaveManager = async (host, port, packageId, args) => {

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

    console.log(request)
    return enclaveManagerClient.runStarlarkPackage(request);
}
