import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { Buffer } from "buffer";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { KURTOSIS_EM_API_DEFAULT_URL } from "../constants";
import { KurtosisClient } from "./KurtosisClient";

export class AuthenticatedLocalKurtosisClient extends KurtosisClient {
  private readonly username: string;
  private readonly password: string;

  constructor(username: string, password: string) {
    const defaultUrl = new URL(`${window.location.protocol}//${window.location.host}`);
    var baseUrl = KURTOSIS_EM_API_DEFAULT_URL;
    if (window.env !== undefined && window.env.domain !== undefined) {
      baseUrl = "https://" + window.env.domain;
    }
    super(
      createPromiseClient(KurtosisEnclaveManagerServer, createConnectTransport({ baseUrl: baseUrl })),
      defaultUrl,
      defaultUrl,
    );
    this.username = username;
    this.password = password;
  }

  getHeaderOptions(): { headers?: Headers } {
    const headers = new Headers();
    headers.set("Authorization", `Basic ${Buffer.from(this.username + ":" + this.password).toString("base64")}`);
    return { headers: headers };
  }

  isRunningInCloud(): boolean {
    return false;
  }
}
