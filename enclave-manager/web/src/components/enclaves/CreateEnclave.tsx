import { useCallback, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { KURTOSIS_CREATE_ENCLAVE_URL_ARG } from "../constants";
import { ConfigureEnclaveModal } from "./modals/ConfigureEnclaveModal";
import { ManualCreateEnclaveModal } from "./modals/ManualCreateEnclaveModal";
import { PreloadPackage } from "./PreloadPackage";

export const CreateEnclave = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const [configureEnclaveOpen, setConfigureEnclaveOpen] = useState(false);
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();
  const [manualCreateEnclaveOpen, setManualCreateEnclaveOpen] = useState(false);

  useEffect(() => {
    setManualCreateEnclaveOpen(location.hash === `#${KURTOSIS_CREATE_ENCLAVE_URL_ARG}`);
  }, [location]);

  const handleManualCreateEnclaveConfirmed = (kurtosisPackage: KurtosisPackage) => {
    setKurtosisPackage(kurtosisPackage);
    setManualCreateEnclaveOpen(false);
    setConfigureEnclaveOpen(true);
  };

  const handleOnPackageLoaded = useCallback((kurtosisPackage: KurtosisPackage) => {
    setKurtosisPackage(kurtosisPackage);
    setConfigureEnclaveOpen(true);
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
      <ManualCreateEnclaveModal
        isOpen={manualCreateEnclaveOpen}
        onClose={handleCloseManualCreateEnclave}
        onConfirm={handleManualCreateEnclaveConfirmed}
      />
      {isDefined(kurtosisPackage) && (
        <ConfigureEnclaveModal
          isOpen={configureEnclaveOpen}
          onClose={() => setConfigureEnclaveOpen(false)}
          kurtosisPackage={kurtosisPackage}
        />
      )}
    </>
  );
};
