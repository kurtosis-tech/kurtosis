import { KurtosisClient } from "./KurtosisClient";
import { KURTOSIS_DEFAULT_PORT } from "./constants";
import { createPromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { createConnectTransport } from "@connectrpc/connect-web";

function constructGatewayURL(host: string): string {
  return `https://cloud.kurtosis.com/gateway/ips/${host}/ports/${KURTOSIS_DEFAULT_PORT}`;
}

export class AuthenticatedKurtosisClient extends KurtosisClient {
  private token: string;

  constructor(host: string, token: string) {
    super(
      createPromiseClient(KurtosisEnclaveManagerServer, createConnectTransport({ baseUrl: constructGatewayURL(host) })),
    );
    this.token = token;
  }

  getHeaderOptions(): { headers?: Headers } {
    const headers = new Headers();
    headers.set("Authorization", `Bearer ${this.token}`);
    return { headers: headers };
  }
}
