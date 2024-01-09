import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { paths } from "kurtosis-sdk/src/engine/rest_api_bindings/types";
import createClient from "openapi-fetch";
import {
  KURTOSIS_EM_API_DEFAULT_URL,
  KURTOSIS_REST_API_DEFAULT_URL,
  KURTOSIS_WEBSOCKET_API_DEFAULT_URL,
} from "../constants";
import { KurtosisClient } from "./KurtosisClient";
import { createWSClient } from "./websocketClient/WebSocketClient";

export class LocalKurtosisClient extends KurtosisClient {
  constructor() {
    const defaultUrl = new URL(`${window.location.protocol}//${window.location.host}`);
    super(
      createPromiseClient(
        KurtosisEnclaveManagerServer,
        createConnectTransport({ baseUrl: KURTOSIS_EM_API_DEFAULT_URL }),
      ),
      createClient<paths>({ baseUrl: KURTOSIS_REST_API_DEFAULT_URL }),
      createWSClient<paths>({ baseUrl: KURTOSIS_WEBSOCKET_API_DEFAULT_URL }),
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
