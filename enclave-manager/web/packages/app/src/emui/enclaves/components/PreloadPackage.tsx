import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useSearchParams } from "react-router-dom";
import { KURTOSIS_PACKAGE_ID_URL_ARG } from "./configuration/drawer/constants";
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
