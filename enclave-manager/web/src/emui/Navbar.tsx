import { FiHome } from "react-icons/fi";
import { Link, useLocation } from "react-router-dom";
import { NavButton, Navigation } from "../components/Navigation";

export type NavbarProps = {
  baseApplicationUrl: URL;
};

export const Navbar = ({ baseApplicationUrl }: NavbarProps) => {
  const location = useLocation();
  // const kurtosisClient = useKurtosisClient();

  return (
    <Navigation baseApplicationUrl={baseApplicationUrl}>
      <Link to={"/"}>
        <NavButton
          label={"View enclaves"}
          Icon={<FiHome />}
          isActive={location.pathname === "/" || location.pathname.startsWith("/enclave")}
        />
      </Link>
      {/*{kurtosisClient.isRunningInCloud() && (*/}
      {/*  <Link to={KURTOSIS_CLOUD_CONNECT_URL}>*/}
      {/*    <NavButton label={"Link your CLI"} Icon={<PiLinkSimpleBold />} isActive={true} />*/}
      {/*  </Link>*/}
      {/*)}*/}
      {/*<Link to={"/catalog"}>*/}
      {/*  <NavButton label={"View catalog"} Icon={<FiPackage />} isActive={location.pathname.startsWith("/catalog")} />*/}
      {/*</Link>*/}
    </Navigation>
  );
};
