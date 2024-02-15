import { Box, Flex, IconButton, IconButtonProps, Image, Tooltip } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { useHref } from "react-router-dom";

export type NavigationProps = {};

export const Navigation = ({ children }: PropsWithChildren & NavigationProps) => {
  const logoHref = useHref("/logo.png");

  return (
    <Flex
      as={"nav"}
      className={"primaryNav"}
      flexDirection={"column"}
      alignItems={"center"}
      gap={"36px"}
      position={"fixed"}
      top={"0"}
      h={"100vh"}
      p={"20px 16px"}
    >
      <Flex width={"40px"} height={"40px"} alignItems={"center"}>
        <Image src={logoHref} />
      </Flex>
      <Flex flexDirection={"column"} gap={"16px"} flex={"1"}>
        {children}
      </Flex>
    </Flex>
  );
};

export const NavigationDivider = () => {
  return <Box flex={"1"} />;
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
