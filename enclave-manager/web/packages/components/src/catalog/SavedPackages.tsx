import { createContext, PropsWithChildren, useContext } from "react";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";

type SavedPackagesState = {
  savedPackages: KurtosisPackage[];
  togglePackageSaved: (kurtosisPackage: KurtosisPackage) => void;
};

const SavedPackagesContext = createContext<SavedPackagesState>(null as any);

export const SavedPackagesProvider = ({
  savedPackages,
  togglePackageSaved,
  children,
}: PropsWithChildren<SavedPackagesState>) => {
  return (
    <SavedPackagesContext.Provider value={{ savedPackages, togglePackageSaved }}>
      {children}
    </SavedPackagesContext.Provider>
  );
};

export const useSavedPackages = () => {
  return useContext(SavedPackagesContext);
};
