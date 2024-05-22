import { isDefined } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { KURTOSIS_BUILD_ENCLAVE_URL_ARG } from "./configuration/drawer/constants";
import { EnclaveBuilderDrawer } from "./enclaveBuilder/EnclaveBuilderDrawer";

export const BuildEnclave = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const [buildEnclaveOpen, setBuildEnclaveOpen] = useState(false);

  useEffect(() => {
    setBuildEnclaveOpen(location.hash === `#${KURTOSIS_BUILD_ENCLAVE_URL_ARG}`);
  }, [location]);

  useEffect(() => {
    const query = new URLSearchParams(location.search);
    const code = query.get('code');
    if (code) {
      window.opener.postMessage({ code }, 'http://localhost:4000');
      window.close();
    }
  }, []);

  const handleCloseBuildEnclave = () => {
    setBuildEnclaveOpen(false);
    if (isDefined(location.hash)) {
      navigate(`${location.pathname}${location.search}`);
    }
  };

  return <EnclaveBuilderDrawer isOpen={buildEnclaveOpen} onClose={handleCloseBuildEnclave} />;
};
