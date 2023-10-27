import { createPromiseClient, PromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { asyncResult } from "../../utils";
import { KURTOSIS_CLOUD_URL } from "../constants";
import { KurtosisPackageIndexer } from "./api/kurtosis_package_indexer_connect";
import { ReadPackageRequest } from "./api/kurtosis_package_indexer_pb";

export class KurtosisPackageIndexerClient {
  private client: PromiseClient<typeof KurtosisPackageIndexer>;

  constructor() {
    this.client = createPromiseClient(KurtosisPackageIndexer, createConnectTransport({ baseUrl: KURTOSIS_CLOUD_URL }));
  }

  getPackages = async () => {
    return asyncResult(() => {
      return this.client.getPackages({});
    });
  };

  readPackage = async (packageUrl: string) => {
    return asyncResult(() => {
      const components = packageUrl.split("/");
      if (components.length < 3) {
        throw `Illegal url, invalid number of components: ${packageUrl}`;
      }
      if (components[1].length < 1 || components[2].length < 1) {
        throw `Illegal url, empty components: ${packageUrl}`;
      }
      return this.client.readPackage(
        new ReadPackageRequest({
          repositoryMetadata: {
            baseUrl: "github.com",
            owner: components[1],
            name: components[2],
          },
        }),
      );
    });
  };
}
