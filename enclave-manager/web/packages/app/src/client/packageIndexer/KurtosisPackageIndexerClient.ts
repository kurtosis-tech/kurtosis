import { createPromiseClient, PromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { asyncResult } from "../../utils";
import { parsePackageUrl } from "../../utils/packageUtils";
import { KURTOSIS_PACKAGE_INDEXER_URL } from "../constants";
import { KurtosisPackageIndexer } from "./api/kurtosis_package_indexer_connect";
import { PackageRepository, ReadPackageRequest } from "./api/kurtosis_package_indexer_pb";

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
