import { KurtosisClient } from "./KurtosisClient";
import { KURTOSIS_DEFAULT_URL } from "./constants";
import { createPromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { createConnectTransport } from "@connectrpc/connect-web";

export class LocalKurtosisClient extends KurtosisClient {
  constructor() {
    super(createPromiseClient(KurtosisEnclaveManagerServer, createConnectTransport({ baseUrl: KURTOSIS_DEFAULT_URL })));
  }

  getHeaderOptions() {
    return {};
  }
}
