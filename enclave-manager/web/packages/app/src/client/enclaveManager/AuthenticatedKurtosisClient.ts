import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { paths } from "kurtosis-sdk/src/engine/rest_api_bindings/types";
import { DateTime } from "luxon";
import createClient from "openapi-fetch";
import {
  KURTOSIS_CLOUD_EM_URL,
  KURTOSIS_CLOUD_UI_URL,
  KURTOSIS_DEFAULT_EM_API_PORT,
  KURTOSIS_DEFAULT_REST_API_PORT,
} from "../constants";
import { KurtosisClient } from "./KurtosisClient";
import createWSClient from "./websocketClient/WebSocketClient";

function constructGatewayURL(remoteHost: string): string {
  return `${KURTOSIS_CLOUD_UI_URL}/gateway/ips/${remoteHost}/ports/${KURTOSIS_DEFAULT_EM_API_PORT}`;
}

function constructRESTGatewayURL(remoteHost: string): string {
  return `${KURTOSIS_CLOUD_UI_URL}/gateway/ips/${remoteHost}/ports/${KURTOSIS_DEFAULT_REST_API_PORT}`;
}

function constructWSGatewayURL(remoteHost: string): string {
  return constructRESTGatewayURL(remoteHost).replace(/^http/, 'ws');
}

export class AuthenticatedKurtosisClient extends KurtosisClient {
  private readonly token: string;
  private readonly tokenExpiry: DateTime;

  constructor(gatewayHost: string, token: string, parentUrl: URL, childUrl: URL) {
    super(
      createPromiseClient(
        KurtosisEnclaveManagerServer,
        createConnectTransport({ baseUrl: constructGatewayURL(gatewayHost) }),
      ),
      createClient<paths>({ baseUrl: constructRESTGatewayURL(gatewayHost) }),
      createWSClient<paths>({ baseUrl: constructWSGatewayURL(gatewayHost) }),
      parentUrl,
      childUrl,
    );
    this.token = token;
    const parsedToken = JSON.parse(atob(this.token.split(".")[1]));
    this.tokenExpiry = DateTime.fromSeconds(parsedToken["exp"]);
  }

  validateTokenStillFresh() {
    if (this.tokenExpiry < DateTime.now()) {
      console.log("Token has expired. Triggering a refresh");
      window.location.href = KURTOSIS_CLOUD_EM_URL;
    }
  }

  getHeaderOptions(): { headers?: Headers } {
    this.validateTokenStillFresh();
    const headers = new Headers();
    headers.set("Authorization", `Bearer ${this.token}`);
    return { headers: headers };
  }
  isRunningInCloud(): boolean {
    return true;
  }
}
