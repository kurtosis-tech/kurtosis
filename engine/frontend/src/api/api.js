import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {
    GetListFilesArtifactNamesAndUuidsRequest,
    GetServicesRequest,
    InspectFilesArtifactContentsRequest,
    RunStarlarkPackageRequest,
    GetStarlarkRunRequest,
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
    const request = new GetStarlarkRunRequest(
        {
            "apicIpAddress": host,
            "apicPort": port,
        }
    );
    // return enclaveManagerClient.getStarlarkRun(request, createHeaderOptionsWithToken(token));
    return rawData;
}

const rawData =
    {
        "experimental_features": [],
        "package_id": "github.com/kurtosis-tech/etcd-package",
        "serialized_script": "NAME_ARG = \"etcd_name\"\nNAME_ARG_DEFAULT = \"etcd\"\n\nIMAGE_ARG = \"etcd_image\"\nIMAGE_ARG_DEFAULT = \"softlang/etcd-alpine:v3.4.14\"\n\nCLIENT_PORT_ARG = \"etcd_client_port\"\nCLIENT_PORT_ARG_DEFAULT = 2379\n\nENV_VARS_ARG = \"etcd_env_vars\"\nENV_VARS_ARG_DEFAULT = {}\n\ndef run(plan, args):\n\n    name = args.get(NAME_ARG, NAME_ARG_DEFAULT)\n    image = args.get(IMAGE_ARG, IMAGE_ARG_DEFAULT)\n    client_port = args.get(CLIENT_PORT_ARG, CLIENT_PORT_ARG_DEFAULT)\n    env_vars_overrides = args.get(ENV_VARS_ARG, ENV_VARS_ARG_DEFAULT)\n    env_vars = {\n        \"ALLOW_NONE_AUTHENTICATION\": \"yes\",\n        \"ETCD_DATA_DIR\": \"/etcd_data\",\n        \"ETCD_LISTEN_CLIENT_URLS\": \"http://0.0.0.0:{}\".format(client_port),\n        \"ETCD_ADVERTISE_CLIENT_URLS\": \"http://0.0.0.0:{}\".format(client_port),\n    } | env_vars_overrides\n\n    etcd_service_config= ServiceConfig(\n        image = image,\n        ports = {\n            \"client\": PortSpec(number = client_port, transport_protocol = \"TCP\")\n        },\n        env_vars = env_vars,\n        ready_conditions = ReadyCondition(\n            recipe = ExecRecipe(\n                command = [\"etcdctl\", \"get\", \"test\"]\n            ),\n            field = \"code\",\n            assertion = \"==\",\n            target_value = 0\n        )\n    )\n\n    etcd = plan.add_service(name = name, config = etcd_service_config)\n\n    return {\"service-name\": name, \"hostname\": etcd.hostname, \"port\": client_port}\n\n",
        "serialized_params": "{}",
        "parallelism": 4,
        "relative_path_to_main_file": "main.star",
        "main_function_name": "",
        "is_production": false
    }



