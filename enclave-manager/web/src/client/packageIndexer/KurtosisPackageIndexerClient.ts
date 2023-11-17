import { createPromiseClient, PromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { asyncResult } from "../../utils";
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

  parsePackageUrl(packageUrl: string) {
    const components = packageUrl.split("/");
    if (components.length < 3) {
      throw Error(`Illegal url, invalid number of components: ${packageUrl}`);
    }
    if (components[1].length < 1 || components[2].length < 1) {
      throw Error(`Illegal url, empty components: ${packageUrl}`);
    }
    return new PackageRepository({
      baseUrl: "github.com",
      owner: components[1],
      name: components[2],
      rootPath: components.filter((v, i) => i > 2 && v.length > 0).join("/") + "/",
    });
  }

  readPackage = async (packageUrl: string) => {
    return asyncResult(() => {
      return this.client.readPackage(new ReadPackageRequest({ repositoryMetadata: this.parsePackageUrl(packageUrl) }));
    });
  };
}
