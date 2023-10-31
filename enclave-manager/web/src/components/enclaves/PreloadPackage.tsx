import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";
import {KURTOSIS_PACKAGE_NAME_URL_ARG} from "../constants";

type PreloadEnclaveProps = {
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

export const PreloadPackage = ({ onPackageLoaded }: PreloadEnclaveProps) => {
  const searchParams = new URLSearchParams(window.location.search);
  const preloadPackage = searchParams.get(KURTOSIS_PACKAGE_NAME_URL_ARG);

  if (!isDefined(preloadPackage)) {
    return null;
  }

  return <PackageLoadingModal packageId={preloadPackage} onPackageLoaded={onPackageLoaded} />;
};
