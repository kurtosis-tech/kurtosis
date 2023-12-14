import { NavButton, Navigation } from "kurtosis-ui-components";
import { FiHome, FiPackage } from "react-icons/fi";
import { PiLinkSimpleBold } from "react-icons/pi";
import { Link, useLocation } from "react-router-dom";
import { KURTOSIS_CLOUD_CONNECT_URL } from "../client/constants";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";

export const Navbar = () => {
  const location = useLocation();
  const kurtosisClient = useKurtosisClient();

  return (
    <Navigation>
      <Link to={"/"}>
        <NavButton
          label={"View enclaves"}
          Icon={<FiHome />}
          isActive={location.pathname === "/" || location.pathname.startsWith("/enclave")}
        />
      </Link>
      <Link to={"/catalog"}>
        <NavButton label={"View catalog"} Icon={<FiPackage />} isActive={location.pathname.startsWith("/catalog")} />
      </Link>
      {kurtosisClient.isRunningInCloud() && (
        <Link to={KURTOSIS_CLOUD_CONNECT_URL}>
          <NavButton label={"Link your CLI"} Icon={<PiLinkSimpleBold />} />
        </Link>
      )}
    </Navigation>
  );
};
