import { PromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";

export abstract class KurtosisClient {
  protected client: PromiseClient<typeof KurtosisEnclaveManagerServer>;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>) {
    this.client = client;
  }

  abstract getHeaderOptions(): { headers?: Headers };

  async getEnclaves() {
    return this.client.getEnclaves({}, this.getHeaderOptions());
  }
}
