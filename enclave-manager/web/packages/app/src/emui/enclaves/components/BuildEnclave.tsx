import { isDefined } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useSettings } from "../../settings";
import { KURTOSIS_BUILD_ENCLAVE_URL_ARG } from "./configuration/drawer/constants";
import { EnclaveBuilderDrawer } from "./enclaveBuilder/EnclaveBuilderDrawer";

export const BuildEnclave = () => {
  const { settings } = useSettings();
  const navigate = useNavigate();
  const location = useLocation();

  const [buildEnclaveOpen, setBuildEnclaveOpen] = useState(false);

  useEffect(() => {
    setBuildEnclaveOpen(location.hash === `#${KURTOSIS_BUILD_ENCLAVE_URL_ARG}`);
  }, [location]);

  const handleCloseBuildEnclave = () => {
    setBuildEnclaveOpen(false);
    if (isDefined(location.hash)) {
      navigate(`${location.pathname}${location.search}`);
    }
  };

  if (!settings.ENABLE_EXPERIMENTAL_BUILD_ENCLAVE) {
    return null;
  }

  return <EnclaveBuilderDrawer isOpen={buildEnclaveOpen} onClose={handleCloseBuildEnclave} />;
};
