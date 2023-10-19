import { Flex, IconButton, Tooltip, useColorModeValue } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

export const Navigation = ({ children }: PropsWithChildren) => {
  return (
    <Flex
      flexDirection={"column"}
      gap={"1rem"}
      bg={useColorModeValue("white", "gray.900")}
      borderRight="1px"
      borderRightColor={useColorModeValue("gray.200", "gray.700")}
      h={"100vh"}
      p={"1rem"}
    >
      {children}
    </Flex>
  );
};

type NavButtonProps = {
  label: string;
  Icon: React.ReactElement;
};

export const NavButton = ({ Icon, label }: NavButtonProps) => {
  return (
    <Tooltip label={label} hasArrow placement={"right"} openDelay={500}>
      <IconButton
        aria-label={label}
        variant={"ghost"}
        size={"lg"}
        icon={Icon}
      />
    </Tooltip>
  );
};
