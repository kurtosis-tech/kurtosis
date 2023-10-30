import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { KURTOSIS_DEFAULT_URL } from "../constants";
import { KurtosisClient } from "./KurtosisClient";

export class LocalKurtosisClient extends KurtosisClient {
  constructor() {
    super(createPromiseClient(KurtosisEnclaveManagerServer, createConnectTransport({ baseUrl: KURTOSIS_DEFAULT_URL })));
  }

  getHeaderOptions() {
    return {};
  }
}
