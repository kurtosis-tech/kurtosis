import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useCallback, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { KURTOSIS_CREATE_ENCLAVE_URL_ARG } from "./configuration/drawer/constants";
import { CreateOrConfigureEnclaveDrawer } from "./configuration/drawer/CreateOrConfigureEnclaveDrawer";
import { PreloadPackage } from "./PreloadPackage";

export const CreateEnclave = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();
  const [manualCreateEnclaveOpen, setManualCreateEnclaveOpen] = useState(false);

  useEffect(() => {
    setManualCreateEnclaveOpen(location.hash === `#${KURTOSIS_CREATE_ENCLAVE_URL_ARG}`);
  }, [location]);

  const handleOnPackageLoaded = useCallback((kurtosisPackage: KurtosisPackage) => {
    setKurtosisPackage(kurtosisPackage);
    setManualCreateEnclaveOpen(true);
  }, []);

  const handleCloseManualCreateEnclave = () => {
    setManualCreateEnclaveOpen(false);
    if (isDefined(location.hash)) {
      navigate(`${location.pathname}${location.search}`);
    }
  };

  return (
    <>
      <PreloadPackage onPackageLoaded={handleOnPackageLoaded} />
      <CreateOrConfigureEnclaveDrawer
        isOpen={manualCreateEnclaveOpen}
        onClose={handleCloseManualCreateEnclave}
        kurtosisPackage={kurtosisPackage}
      />
    </>
  );
};
