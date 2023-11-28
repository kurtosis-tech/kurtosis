import { Spinner } from "@chakra-ui/react";
import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useState } from "react";
import { Result } from "true-myth";
import { GetPackagesResponse, KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useKurtosisPackageIndexerClient } from "../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined } from "../../utils";

export type CatalogsState = {
  catalog: Result<GetPackagesResponse, string>;
  refreshCatalog: () => Promise<Result<GetPackagesResponse, string>>;
};

const CatalogContext = createContext<CatalogsState>(null as any);

export const CatalogContextProvider = ({ children }: PropsWithChildren) => {
  const packageIndexerClient = useKurtosisPackageIndexerClient();
  const [catalog, setCatalog] = useState<Result<GetPackagesResponse, string>>();

  const refreshCatalog = useCallback(async () => {
    setCatalog(undefined);
    const catalog = await packageIndexerClient.getPackages();
    setCatalog(catalog);
    return catalog;
  }, [packageIndexerClient]);

  useEffect(() => {
    refreshCatalog();
  }, [refreshCatalog]);

  if (!isDefined(catalog)) {
    return <Spinner />;
  }

  return <CatalogContext.Provider value={{ catalog, refreshCatalog }}>{children}</CatalogContext.Provider>;
};

export const usePackageCatalog = () => {
  const { catalog } = useContext(CatalogContext);
  return catalog;
};

export const useKurtosisPackage = (packageId: string): Result<KurtosisPackage, string> => {
  const catalog = usePackageCatalog();
  const kurtosisPackage = catalog.map((catalog) =>
    catalog.packages.find((kurtosisPackage) => kurtosisPackage.name === packageId),
  );

  if (kurtosisPackage.isErr) {
    return kurtosisPackage.cast<KurtosisPackage>();
  } else {
    if (!isDefined(kurtosisPackage.value)) {
      return Result.err(`No package with id ${packageId} could be found.`);
    }
    return Result.ok(kurtosisPackage.value);
  }
};
