import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";
import {GetServicesRequest} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";

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
            "apicIpAddress": "127.0.0.1",
            "apicPort": 55296,
        }
    );
    return enclaveManagerClient.getServices(request);
}
