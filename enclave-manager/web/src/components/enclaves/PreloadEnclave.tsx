import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";

type PreloadEnclaveProps = {
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

export const PreloadEnclave = ({ onPackageLoaded }: PreloadEnclaveProps) => {
  const searchParams = new URLSearchParams(window.location.search);
  const preloadPackage = searchParams.get("preloadPackage");

  if (!isDefined(preloadPackage)) {
    return null;
  }

  return <PackageLoadingModal packageId={preloadPackage} onPackageLoaded={onPackageLoaded} />;
};
