import { FiHome, FiPackage } from "react-icons/fi";
import { Link, useLocation } from "react-router-dom";
import { NavButton, Navigation } from "../components/Navigation";

export type NavbarProps ={
  baseApplicationUrl: URL
}

export const Navbar = ({ baseApplicationUrl}: NavbarProps) => {
  const location = useLocation();

  return (
    <Navigation baseApplicationUrl={baseApplicationUrl}>
      <Link to={"/"}>
        <NavButton
          label={"View enclaves"}
          Icon={<FiHome />}
          isActive={location.pathname === "/" || location.pathname.startsWith("/enclave")}
        />
      </Link>
      {/*<Link to={"/catalog"}>*/}
      {/*  <NavButton label={"View catalog"} Icon={<FiPackage />} isActive={location.pathname.startsWith("/catalog")} />*/}
      {/*</Link>*/}
    </Navigation>
  );
};
