import { PromiseClient } from "@connectrpc/connect";
import { RunStarlarkPackageArgs } from "enclave-manager-sdk/build/api_container_service_pb";
import {
  CreateEnclaveArgs,
  DestroyEnclaveArgs,
  EnclaveAPIContainerInfo,
  EnclaveInfo,
  EnclaveMode
} from "enclave-manager-sdk/build/engine_service_pb";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {
  GetListFilesArtifactNamesAndUuidsRequest,
  GetServicesRequest,
  GetStarlarkRunRequest,
  RunStarlarkPackageRequest
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import { assertDefined, asyncResult } from "../../utils";
import { RemoveFunctions } from "../../utils/types";

export abstract class KurtosisClient {
  protected readonly client: PromiseClient<typeof KurtosisEnclaveManagerServer>;
  protected readonly browserBasePathUrl: string;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>, browserBasePathUrl: string) {
    this.client = client;
    this.browserBasePathUrl = browserBasePathUrl;
  }

  abstract getHeaderOptions(): { headers?: Headers };

  getBrowserBasePathUrl() {
    return this.browserBasePathUrl;
  }

  async checkHealth() {
    return asyncResult(this.client.check({}, this.getHeaderOptions()));
  }

  async getEnclaves() {
    return asyncResult(this.client.getEnclaves({}, this.getHeaderOptions()), "KurtosisClient could not getEnclaves");
  }

  async destroy(enclaveUUID: string) {
    return asyncResult(
      this.client.destroyEnclave(new DestroyEnclaveArgs({ enclaveIdentifier: enclaveUUID }), this.getHeaderOptions()),
      `KurtosisClient could not destroy enclave ${enclaveUUID}`
    );
  }

  async getServices(enclave: RemoveFunctions<EnclaveInfo>) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(apicInfo, `Cannot getServices because the passed enclave '${enclave.name}' does not have apicInfo`);
      const request = new GetServicesRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave
      });
      return this.client.getServices(request, this.getHeaderOptions());
    }, "KurtosisClient could not getServices");
  }

  async getStarlarkRun(enclave: RemoveFunctions<EnclaveInfo>) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(
        apicInfo,
        `Cannot getStarlarkRun because the passed enclave '${enclave.name}' does not have apicInfo`
      );
      const request = new GetStarlarkRunRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave
      });
      return this.client.getStarlarkRun(request, this.getHeaderOptions());
    }, "KurtosisClient could not getStarlarkRun");
  }

  async listFilesArtifactNamesAndUuids(enclave: RemoveFunctions<EnclaveInfo>) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(
        apicInfo,
        `Cannot listFilesArtifactNamesAndUuids because the passed enclave '${enclave.name}' does not have apicInfo`
      );
      const request = new GetListFilesArtifactNamesAndUuidsRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave
      });
      return this.client.listFilesArtifactNamesAndUuids(request, this.getHeaderOptions());
    }, "KurtosisClient could not listFilesArtifactNamesAndUuids");
  }

  async createEnclave(
    enclaveName: string,
    apiContainerLogLevel: string,
    productionMode?: boolean,
    apiContainerVersionTag?: string
  ) {
    return asyncResult(() => {
      const request = new CreateEnclaveArgs({
        enclaveName,
        apiContainerLogLevel,
        mode: productionMode ? EnclaveMode.PRODUCTION : EnclaveMode.TEST,
        apiContainerVersionTag: apiContainerVersionTag || ""
      });
      return this.client.createEnclave(request, this.getHeaderOptions());
    });
  }

  async runStarlarkPackage(
    apicInfo: RemoveFunctions<EnclaveAPIContainerInfo>,
    packageId: string,
    args: Record<string, any>
  ) {
    // Not currently using asyncResult as the return type here is an asyncIterable
    const request = new RunStarlarkPackageRequest({
      apicIpAddress: apicInfo.bridgeIpAddress,
      apicPort: apicInfo.grpcPortInsideEnclave,
      RunStarlarkPackageArgs: new RunStarlarkPackageArgs({
        dryRun: false,
        packageId: packageId,
        serializedParams: JSON.stringify(args)
      })
    });
    return this.client.runStarlarkPackage(request, this.getHeaderOptions());
  }
}
