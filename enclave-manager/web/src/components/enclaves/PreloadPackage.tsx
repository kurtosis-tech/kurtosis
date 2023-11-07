import { useSearchParams } from "react-router-dom";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { KURTOSIS_PACKAGE_ID_URL_ARG } from "../constants";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";

type PreloadEnclaveProps = {
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

export const PreloadPackage = ({ onPackageLoaded }: PreloadEnclaveProps) => {
  const [searchParams] = useSearchParams();
  const packageId = searchParams.get(KURTOSIS_PACKAGE_ID_URL_ARG);

  if (!isDefined(packageId)) {
    return null;
  }

  return <PackageLoadingModal packageId={packageId} onPackageLoaded={onPackageLoaded} />;
};
