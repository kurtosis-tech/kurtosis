import { useCallback, useEffect, useState } from "react";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { PreloadPackage } from "./PreloadPackage";
import { ManualCreateEnclaveModal } from "./modals/ManualCreateEnclaveModal";
import { isDefined } from "../../utils";
import { ConfigureEnclaveModal } from "./modals/ConfigureEnclaveModal";
import { KURTOSIS_CREATE_ENCLAVE_URL_ARG } from "../constants";
import { useLocation, useNavigate } from "react-router-dom";

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

  const handleOnPackageLoaded = useCallback(
    (kurtosisPackage: KurtosisPackage) => {
      setKurtosisPackage(kurtosisPackage);
      setConfigureEnclaveOpen(true);
    }, []);

  const handleCloseManualCreateEnclave = () => {
    setManualCreateEnclaveOpen(false);
    if (location.hash) {
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
