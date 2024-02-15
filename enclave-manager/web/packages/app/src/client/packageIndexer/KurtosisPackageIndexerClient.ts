import { createPromiseClient, PromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { KurtosisPackageIndexer, PackageRepository, ReadPackageRequest } from "kurtosis-cloud-indexer-sdk";
import { asyncResult, parsePackageUrl } from "kurtosis-ui-components";
import { KURTOSIS_PACKAGE_INDEXER_URL } from "../constants";

export class KurtosisPackageIndexerClient {
  private client: PromiseClient<typeof KurtosisPackageIndexer>;

  constructor() {
    this.client = createPromiseClient(
      KurtosisPackageIndexer,
      createConnectTransport({ baseUrl: KURTOSIS_PACKAGE_INDEXER_URL }),
    );
  }

  getPackages = async () => {
    return asyncResult(() => {
      return this.client.getPackages({});
    });
  };

  readPackage = async (packageUrl: string) => {
    return asyncResult(() => {
      return this.client.readPackage(
        new ReadPackageRequest({ repositoryMetadata: new PackageRepository(parsePackageUrl(packageUrl)) }),
      );
    });
  };
}
