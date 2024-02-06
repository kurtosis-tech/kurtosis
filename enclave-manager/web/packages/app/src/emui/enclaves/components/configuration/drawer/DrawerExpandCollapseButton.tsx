import { IconButton } from "@chakra-ui/react";
import { BiArrowToLeft, BiArrowToRight } from "react-icons/bi";
import { DrawerSizes } from "./types";

type DrawerExpandCollapseButtonProps = {
  drawerSize: DrawerSizes;
  onClick: () => void;
};
export const DrawerExpandCollapseButton = ({ drawerSize, onClick }: DrawerExpandCollapseButtonProps) => {
  return (
    <IconButton
      size={"sm"}
      icon={drawerSize === "xl" ? <BiArrowToLeft /> : <BiArrowToRight />}
      aria-label={"Expand/collapse"}
      variant={"ghost"}
      onClick={onClick}
    />
  );
};
