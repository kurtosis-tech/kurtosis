import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { KURTOSIS_CLOUD_UI_URL, KURTOSIS_DEFAULT_EM_API_PORT } from "../constants";
import { KurtosisClient } from "./KurtosisClient";

function constructGatewayURL(remoteHost: string): string {
  return `${KURTOSIS_CLOUD_UI_URL}/gateway/ips/${remoteHost}/ports/${KURTOSIS_DEFAULT_EM_API_PORT}`;
}

export class AuthenticatedKurtosisClient extends KurtosisClient {
  private readonly token: string;

  constructor(gatewayHost: string, token: string, parentUrl: URL, childUrl: URL) {
    super(
      createPromiseClient(
        KurtosisEnclaveManagerServer,
        createConnectTransport({ baseUrl: constructGatewayURL(gatewayHost) }),
      ),
      parentUrl,
      childUrl,
    );
    this.token = token;
  }

  getHeaderOptions(): { headers?: Headers } {
    const headers = new Headers();
    headers.set("Authorization", `Bearer ${this.token}`);
    return { headers: headers };
  }
}
