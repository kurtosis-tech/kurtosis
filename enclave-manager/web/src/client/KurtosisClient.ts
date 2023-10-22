import { PromiseClient } from "@connectrpc/connect";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { GetServicesRequest, GetStarlarkRunRequest } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import { assertDefined } from "../utils";

export abstract class KurtosisClient {
  protected client: PromiseClient<typeof KurtosisEnclaveManagerServer>;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>) {
    this.client = client;
  }

  abstract getHeaderOptions(): { headers?: Headers };

  async getEnclaves() {
    return this.client.getEnclaves({}, this.getHeaderOptions());
  }

  async getServices(enclave: EnclaveInfo) {
    console.log(enclave);
    const apicInfo = enclave.apiContainerInfo;
    assertDefined(apicInfo, `Cannot getServices because the passed enclave '${enclave.name}' does not have apicInfo`);
    const request = new GetServicesRequest({
      apicIpAddress: apicInfo.bridgeIpAddress,
      apicPort: apicInfo.grpcPortInsideEnclave,
    });
    return this.client.getServices(request, this.getHeaderOptions());
  }

  async getStarlarkRun(enclave: EnclaveInfo) {
    console.log(enclave);
    const apicInfo = enclave.apiContainerInfo;
    assertDefined(
      apicInfo,
      `Cannot getStarlarkRun because the passed enclave '${enclave.name}' does not have apicInfo`,
    );
    const request = new GetStarlarkRunRequest({
      apicIpAddress: apicInfo.bridgeIpAddress,
      apicPort: apicInfo.grpcPortInsideEnclave,
    });
    return this.client.getStarlarkRun(request, this.getHeaderOptions());
  }
}
