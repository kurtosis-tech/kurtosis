import { FiHome, FiPackage } from "react-icons/fi";
import { Link, useLocation } from "react-router-dom";
import { NavButton, Navigation } from "../components/Navigation";

export const Navbar = () => {
  const location = useLocation();

  return (
    <Navigation>
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
