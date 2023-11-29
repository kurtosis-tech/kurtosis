import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined, stringifyError } from "../../utils";

const SAVED_PACKAGES_LOCAL_STORAGE_KEY = "kurtosis-saved-packages";

export const storeSavedPackages = (kurtosisPackages: KurtosisPackage[]) => {
  localStorage.setItem(
    SAVED_PACKAGES_LOCAL_STORAGE_KEY,
    JSON.stringify(kurtosisPackages.map((kurtosisPackage) => kurtosisPackage.name)),
  );
};

export const loadSavedPackageNames = () => {
  try {
    const savedRawValue = localStorage.getItem(SAVED_PACKAGES_LOCAL_STORAGE_KEY);

    if (!isDefined(savedRawValue)) {
      return [];
    }

    return JSON.parse(savedRawValue);
  } catch (error: any) {
    console.error(`Unable to load saved package names. Got error: ${stringifyError(error)}`);
    return [];
  }
};
