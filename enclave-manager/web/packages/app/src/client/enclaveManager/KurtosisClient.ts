import { PromiseClient } from "@connectrpc/connect";
import {
  DownloadFilesArtifactArgs,
  FilesArtifactNameAndUuid,
  RunStarlarkPackageArgs,
  ServiceInfo,
} from "enclave-manager-sdk/build/api_container_service_pb";
import {
  CreateEnclaveArgs,
  DestroyEnclaveArgs,
  EnclaveAPIContainerInfo,
  EnclaveInfo,
  EnclaveMode,
  GetServiceLogsArgs,
  LogLineFilter,
} from "enclave-manager-sdk/build/engine_service_pb";
import { KurtosisEnclaveManagerServer } from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_connect";
import {
  DownloadFilesArtifactRequest,
  GetListFilesArtifactNamesAndUuidsRequest,
  GetServicesRequest,
  GetStarlarkRunRequest,
  InspectFilesArtifactContentsRequest,
  RunStarlarkPackageRequest,
} from "enclave-manager-sdk/build/kurtosis_enclave_manager_api_pb";
import { paths } from "kurtosis-sdk/src/engine/rest_api_bindings/types";
import { assertDefined, asyncResult, isDefined, RemoveFunctions } from "kurtosis-ui-components";
import createClient from "openapi-fetch";
import { KURTOSIS_WEBSOCKET_API_DEFAULT_URL } from "../../client/constants";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import createWSClient from "./websocketClient/WebSocketClient";

type KurtosisRestClient = ReturnType<typeof createClient<paths>>;

export abstract class KurtosisClient {
  protected readonly client: PromiseClient<typeof KurtosisEnclaveManagerServer>;
  protected readonly restClient: KurtosisRestClient;

  /* Full URL of the browser containing the EM UI covering two use cases:
   * In local-mode this is: http://localhost:9711, http://localhost:3000 (with `yarn start` / dev mode)
   * In authenticated mode this is: https://cloud.kurtosis.com/enclave-manager (this data/url is provided as a search param when the code loads)
   *
   * This URL is primarily used to generate links to the EM UI (where the hostname is included).
   * */
  protected readonly cloudUrl: URL;

  /* Full URL of the EM UI, covering two use cases:
   * In local-mode this is the same as the `parentUrl`
   * In authenticated mode : https://cloud.kurtosis.com/enclave-manager/gateway/ips/1-2-3-4/ports/1234/?searchparams... (this data/url is provided as a search param when the code loads)
   *
   * This URL is primarily used to set the React router basename so that the router is able to ignore any leading subdirectories.
   * */
  protected readonly baseApplicationUrl: URL;

  constructor(
    client: PromiseClient<typeof KurtosisEnclaveManagerServer>,
    restClient: KurtosisRestClient,
    parentUrl: URL,
    childUrl: URL,
  ) {
    this.client = client;
    this.restClient = restClient;
    this.cloudUrl = parentUrl;
    this.baseApplicationUrl = childUrl;
    this.getParentRequestedRoute();
  }

  getParentRequestedRoute() {
    const splits = this.cloudUrl.pathname.split("/enclave-manager");
    if (splits[1]) {
      return splits[1];
    }
    return undefined;
  }

  abstract isRunningInCloud(): boolean;

  abstract getHeaderOptions(): { headers?: Headers };

  getCloudBasePathUrl() {
    return `${this.cloudUrl.origin}${this.cloudUrl.pathname}`;
  }

  getBaseApplicationUrl() {
    return this.baseApplicationUrl;
  }

  async checkHealth() {
    console.log(await this.restClient.GET("/engine/info"));
    return asyncResult(this.client.check({}, this.getHeaderOptions()));
  }

  async ServiceLogs2() {
    var wssClient = createWSClient<paths>({ baseUrl: KURTOSIS_WEBSOCKET_API_DEFAULT_URL });
    wssClient.GET("/enclaves/{enclave_identifier}/services/{service_identifier}/logs", (resp) => console.log(resp), {
      params: {
        path: {
          enclave_identifier: "2c70255d55874b5ea2fd3d84d7c1d5f7",
          service_identifier: "81bce6e8ceb0428594b232a6f1802c91",
        },
        query: { follow_logs: true },
      },
    });
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
    }, `KurtosisClient could not getServices for ${enclave.name}`);
  }

  async getServiceLogs(
    abortController: AbortController,
    enclave: RemoveFunctions<EnclaveFullInfo>,
    services: ServiceInfo[],
    followLogs?: boolean,
    numLogLines?: number,
    returnAllLogs?: boolean,
    conjunctiveFilters: LogLineFilter[] = [],
  ) {
    // Not currently using asyncResult as the return type here is an asyncIterable
    const request = new GetServiceLogsArgs({
      enclaveIdentifier: enclave.name,
      serviceUuidSet: services.reduce((acc, service) => ({ ...acc, [service.serviceUuid]: true }), {}),
      followLogs: isDefined(followLogs) ? followLogs : true,
      conjunctiveFilters: conjunctiveFilters,
      numLogLines: isDefined(numLogLines) ? numLogLines : 1500,
      returnAllLogs: !!returnAllLogs,
    });
    return this.client.getServiceLogs(request, { ...this.getHeaderOptions(), signal: abortController.signal });
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
    }, `KurtosisClient could not getStarlarkRun for ${enclave.name}`);
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
    }, `KurtosisClient could not listFilesArtifactNamesAndUuids for ${enclave.name}`);
  }

  async inspectFilesArtifactContents(enclave: RemoveFunctions<EnclaveInfo>, fileUuid: string) {
    return await asyncResult(() => {
      const apicInfo = enclave.apiContainerInfo;
      assertDefined(
        apicInfo,
        `Cannot inspect files artifact contents because the passed enclave '${enclave.name}' does not have apicInfo`,
      );
      const request = new InspectFilesArtifactContentsRequest({
        apicIpAddress: apicInfo.bridgeIpAddress,
        apicPort: apicInfo.grpcPortInsideEnclave,
        fileNamesAndUuid: { fileUuid },
      });
      return this.client.inspectFilesArtifactContents(request, this.getHeaderOptions());
    }, `KurtosisClient could not inspectFilesArtifactContents for ${enclave.name}`);
  }

  async downloadFilesArtifact(enclave: RemoveFunctions<EnclaveInfo>, file: FilesArtifactNameAndUuid) {
    const apicInfo = enclave.apiContainerInfo;
    assertDefined(
      apicInfo,
      `Cannot download files artifact because the passed enclave '${enclave.name}' does not have apicInfo`,
    );
    // Not currently using asyncResult as the return type here is an asyncIterable
    const request = new DownloadFilesArtifactRequest({
      apicIpAddress: apicInfo.bridgeIpAddress,
      apicPort: apicInfo.grpcPortInsideEnclave,
      downloadFilesArtifactsArgs: new DownloadFilesArtifactArgs({ identifier: file.fileUuid }),
    });
    return this.client.downloadFilesArtifact(request, this.getHeaderOptions());
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
