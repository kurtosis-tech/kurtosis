import { PromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";

export abstract class KurtosisClient {
  protected client;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>) {}
}
