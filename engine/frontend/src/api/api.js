import {KurtosisEnclaveManagerServer} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {createPromiseClient} from "@bufbuild/connect";
import {createConnectTransport,} from "@bufbuild/connect-web";

const transport = createConnectTransport({
    baseUrl: "http://localhost:8081"
})

const enclaveManagerClient = createPromiseClient(KurtosisEnclaveManagerServer, transport);

export const getEnclavesFromEnclaveManager = async () => {
    return enclaveManagerClient.getEnclaves({});
}
