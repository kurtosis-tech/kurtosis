import { Flex, IconButton, IconButtonProps, Tooltip } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

export const Navigation = ({ children }: PropsWithChildren) => {
  return (
    <Flex as={"nav"} className={"primaryNav"} flexDirection={"column"} gap={"1rem"} h={"100vh"} p={"3rem 1rem"}>
      {children}
    </Flex>
  );
};

type NavButtonProps = Omit<IconButtonProps, "aria-label"> & {
  label: string;
  Icon: React.ReactElement;
};

export const NavButton = ({ Icon, label, ...iconButtonProps }: NavButtonProps) => {
  return (
    <Tooltip label={label} hasArrow placement={"right"} openDelay={500}>
      <IconButton
        {...iconButtonProps}
        colorScheme={"kurtosis"}
        aria-label={label}
        variant={"nav"}
        size={"lg"}
        icon={Icon}
      />
    </Tooltip>
  );
};
