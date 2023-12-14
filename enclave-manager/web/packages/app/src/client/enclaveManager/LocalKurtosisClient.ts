import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { KURTOSIS_EM_API_DEFAULT_URL } from "../constants";
import { KurtosisClient } from "./KurtosisClient";

export class LocalKurtosisClient extends KurtosisClient {
  constructor() {
    const defaultUrl = new URL(`${window.location.protocol}//${window.location.host}`);
    super(
      createPromiseClient(
        KurtosisEnclaveManagerServer,
        createConnectTransport({ baseUrl: KURTOSIS_EM_API_DEFAULT_URL }),
      ),
      defaultUrl,
      defaultUrl,
    );
  }

  getHeaderOptions() {
    return {};
  }

  isRunningInCloud(): boolean {
    return false;
  }
}
