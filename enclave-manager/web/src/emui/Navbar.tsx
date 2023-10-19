import { NavButton, Navigation } from "../components/Navigation";
import { FiBookOpen, FiMonitor } from "react-icons/fi";
import React from "react";

export const Navbar = () => {
  return (
    <Navigation>
      <NavButton label={"View enclaves"} Icon={<FiMonitor />} />
      <NavButton label={"View catalog"} Icon={<FiBookOpen />} />
    </Navigation>
  );
};
