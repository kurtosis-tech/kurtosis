import { Flex, Heading, Spinner } from "@chakra-ui/react";
import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useState } from "react";
import { Result } from "true-myth";
import { GetPackagesResponse, KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useKurtosisPackageIndexerClient } from "../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined } from "../../utils";
import { loadSavedPackageNames, storeSavedPackages } from "./storage";

export type CatalogsState = {
  catalog: Result<GetPackagesResponse, string>;
  savedPackages: KurtosisPackage[];
  refreshCatalog: () => Promise<Result<GetPackagesResponse, string>>;
  togglePackageSaved: (kurtosisPackage: KurtosisPackage) => void;
};

const CatalogContext = createContext<CatalogsState>(null as any);

export const CatalogContextProvider = ({ children }: PropsWithChildren) => {
  const packageIndexerClient = useKurtosisPackageIndexerClient();
  const [catalog, setCatalog] = useState<Result<GetPackagesResponse, string>>();
  const [savedPackages, setSavedPackages] = useState<KurtosisPackage[]>([]);

  const refreshCatalog = useCallback(async () => {
    setCatalog(undefined);
    const catalog = await packageIndexerClient.getPackages();
    setCatalog(catalog);

    if (catalog.isOk) {
      const savedPackageNames = new Set(loadSavedPackageNames());
      setSavedPackages(catalog.value.packages.filter((kurtosisPackage) => savedPackageNames.has(kurtosisPackage.name)));
    }

    return catalog;
  }, [packageIndexerClient]);

  const togglePackageSaved = useCallback((kurtosisPackage: KurtosisPackage) => {
    setSavedPackages((savedPackages) => {
      const packageSavedAlready = savedPackages.some((p) => p.name === kurtosisPackage.name);
      const newSavedPackages: KurtosisPackage[] = packageSavedAlready
        ? savedPackages.filter((p) => p.name !== kurtosisPackage.name)
        : [...savedPackages, kurtosisPackage];
      storeSavedPackages(newSavedPackages);
      return newSavedPackages;
    });
  }, []);

  useEffect(() => {
    refreshCatalog();
  }, [refreshCatalog]);

  if (!isDefined(catalog)) {
    return (
      <Flex width="100%" direction="column" alignItems={"center"} gap={"1rem"} padding={"3rem"}>
        <Spinner size={"xl"} />
        <Heading as={"h2"} fontSize={"2xl"}>
          Fetching Catalog...
        </Heading>
      </Flex>
    );
  }

  return (
    <CatalogContext.Provider value={{ catalog, refreshCatalog, togglePackageSaved, savedPackages }}>
      {children}
    </CatalogContext.Provider>
  );
};

export const useCatalogContext = () => {
  return useContext(CatalogContext);
};

export const usePackageCatalog = () => {
  const { catalog } = useCatalogContext();
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
