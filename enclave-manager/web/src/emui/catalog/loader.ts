import { defer } from "react-router-dom";
import { Result } from "true-myth";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { KurtosisPackageIndexerClient } from "../../client/packageIndexer/KurtosisPackageIndexerClient";

const loadCatalog = async (
  kurtosisIndexerClient: KurtosisPackageIndexerClient,
): Promise<Result<KurtosisPackage[], string>> => {
  const packagesResponse = await kurtosisIndexerClient.getPackages();
  if (packagesResponse.isErr) {
    return Result.err(packagesResponse.error || "Unknown api error");
  }

  return Result.ok(packagesResponse.value.packages);
};

export type CatalogLoaderResolved = {
  catalog: Awaited<ReturnType<typeof loadCatalog>>;
};

export const catalogLoader = (kurtosisIndexerClient: KurtosisPackageIndexerClient) => async () => {
  return defer({ catalog: loadCatalog(kurtosisIndexerClient) });
};
