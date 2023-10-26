import { createConnectTransport } from "@connectrpc/connect-web";
import { createPromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "../../build/kurtosis_enclave_manager_api_connect.js";

describe("Enclave manager SDK", () => {
  it("Should be able to be instantiated", () => {
    const transport = createConnectTransport({baseUrl: "someUrl"});
    const client = createPromiseClient(KurtosisEnclaveManagerServer, transport);
    expect(client).toBeDefined();
  })
})