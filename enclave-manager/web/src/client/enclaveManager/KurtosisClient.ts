import { PromiseClient } from "@connectrpc/connect";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {
  GetListFilesArtifactNamesAndUuidsRequest,
  GetServicesRequest,
  GetStarlarkRunRequest,
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import { assertDefined, asyncResult } from "../../utils";

export abstract class KurtosisClient {
  protected client: PromiseClient<typeof KurtosisEnclaveManagerServer>;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>) {
    this.client = client;
  }

  abstract getHeaderOptions(): { headers?: Headers };

  async checkHealth() {
    return asyncResult(this.client.check({}, this.getHeaderOptions()));
  }

  async getEnclaves() {
    return asyncResult(this.client.getEnclaves({}, this.getHeaderOptions()), "KurtosisClient could not getEnclaves");
  }

  async getServices(enclave: EnclaveInfo) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(apicInfo, `Cannot getServices because the passed enclave '${enclave.name}' does not have apicInfo`);
      const request = new GetServicesRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave,
      });
      return this.client.getServices(request, this.getHeaderOptions());
    }, "KurtosisClient could not getServices");
  }

  async getStarlarkRun(enclave: EnclaveInfo) {
    return await asyncResult(() => {
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
    }, "KurtosisClient could not getStarlarkRun");
  }

  async listFilesArtifactNamesAndUuids(enclave: EnclaveInfo) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(
        apicInfo,
        `Cannot listFilesArtifactNamesAndUuids because the passed enclave '${enclave.name}' does not have apicInfo`,
      );
      const request = new GetListFilesArtifactNamesAndUuidsRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave,
      });
      return this.client.listFilesArtifactNamesAndUuids(request, this.getHeaderOptions());
    }, "KurtosisClient could not listFilesArtifactNamesAndUuids");
  }
}
