import { PromiseClient } from "@connectrpc/connect";
import { RunStarlarkPackageArgs } from "enclave-manager-sdk/build/api_container_service_pb";
import {
  CreateEnclaveArgs,
  DestroyEnclaveArgs,
  EnclaveAPIContainerInfo,
  EnclaveInfo,
  EnclaveMode,
} from "enclave-manager-sdk/build/engine_service_pb";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {
  GetListFilesArtifactNamesAndUuidsRequest,
  GetServicesRequest,
  GetStarlarkRunRequest,
  RunStarlarkPackageRequest,
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import { assertDefined, asyncResult } from "../../utils";
import { RemoveFunctions } from "../../utils/types";

export abstract class KurtosisClient {
  protected readonly client: PromiseClient<typeof KurtosisEnclaveManagerServer>;

  /* Full URL of the browser containing the EM UI covering two use cases:
   * In local-mode this is: http://localhost:9711, http://localhost:3000 (with `yarn start` / dev mode)
   * In authenticated mode this is: https://cloud.kurtosis.com/enclave-manager (this data/url is provided as a search param when the code loads)
   *
   * This URL is primarily used to generate links to the EM UI (where the hostname is included).
   * */
  protected readonly parentUrl?: URL;

  /* Full URL of the EM UI, covering two use cases:
   * In local-mode this is the same as the `parentUrl`
   * In authenticated mode : https://cloud.kurtosis.com/enclave-manager/gateway/ips/1-2-3-4/ports/1234/?searchparams... (this data/url is provided as a search param when the code loads)
   *
   * This URL is primarily used to set the react router basename so that the router is able to ignore any leading subdirectories.
   * */
  protected readonly childUrl?: URL;

  constructor(client: PromiseClient<typeof KurtosisEnclaveManagerServer>, parentUrl: URL, childUrl: URL) {
    this.client = client;
    this.parentUrl = parentUrl;
    this.childUrl = childUrl;
  }

  abstract getHeaderOptions(): { headers?: Headers };

  getParentBasePathUrl() {
    return `${this.parentUrl?.origin}${this.parentUrl?.pathname}`;
  }

  getChildPath() {
    return this.childUrl?.pathname;
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
      `KurtosisClient could not destroy enclave ${enclaveUUID}`,
    );
  }

  async getServices(enclave: RemoveFunctions<EnclaveInfo>) {
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

  async getStarlarkRun(enclave: RemoveFunctions<EnclaveInfo>) {
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

  async listFilesArtifactNamesAndUuids(enclave: RemoveFunctions<EnclaveInfo>) {
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

  async createEnclave(
    enclaveName: string,
    apiContainerLogLevel: string,
    productionMode?: boolean,
    apiContainerVersionTag?: string,
  ) {
    return asyncResult(() => {
      const request = new CreateEnclaveArgs({
        enclaveName,
        apiContainerLogLevel,
        mode: productionMode ? EnclaveMode.PRODUCTION : EnclaveMode.TEST,
        apiContainerVersionTag: apiContainerVersionTag || "",
      });
      return this.client.createEnclave(request, this.getHeaderOptions());
    });
  }

  async runStarlarkPackage(
    apicInfo: RemoveFunctions<EnclaveAPIContainerInfo>,
    packageId: string,
    args: Record<string, any>,
  ) {
    // Not currently using asyncResult as the return type here is an asyncIterable
    const request = new RunStarlarkPackageRequest({
      apicIpAddress: apicInfo.bridgeIpAddress,
      apicPort: apicInfo.grpcPortInsideEnclave,
      RunStarlarkPackageArgs: new RunStarlarkPackageArgs({
        dryRun: false,
        packageId: packageId,
        serializedParams: JSON.stringify(args),
      }),
    });
    return this.client.runStarlarkPackage(request, this.getHeaderOptions());
  }
}
