import { FiHome } from "react-icons/fi";
import { BsGear } from "react-icons/bs";
import { Link, useLocation } from "react-router-dom";
import { NavButton, Navigation } from "../components/Navigation";
import { KURTOSIS_CLOUD_CONNECT_URL } from "../client/constants";

export type NavbarProps = {
  baseApplicationUrl: URL;
  isRunningInCloud: boolean;
};

export const Navbar = ({ isRunningInCloud, baseApplicationUrl }: NavbarProps) => {
  const location = useLocation();

  return (
    <Navigation isRunningInCloud={isRunningInCloud} baseApplicationUrl={baseApplicationUrl}>
      <Link to={"/"}>
        <NavButton
          label={"View enclaves"}
          Icon={<FiHome />}
          isActive={location.pathname === "/" || location.pathname.startsWith("/enclave")}
        />
      </Link>
      {isRunningInCloud && (
        <Link to={KURTOSIS_CLOUD_CONNECT_URL}>
          <NavButton
            label={"Link your CLI"}
            Icon={<BsGear />}
            isActive={true}
          />
        </Link>
      )}
      {/*<Link to={"/catalog"}>*/}
      {/*  <NavButton label={"View catalog"} Icon={<FiPackage />} isActive={location.pathname.startsWith("/catalog")} />*/}
      {/*</Link>*/}
    </Navigation>
  );
};
