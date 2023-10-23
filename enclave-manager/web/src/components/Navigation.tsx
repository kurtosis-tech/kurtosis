import { Flex, IconButton, IconButtonProps, Image, Tooltip } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

export const Navigation = ({ children }: PropsWithChildren) => {
  return (
    <Flex
      as={"nav"}
      className={"primaryNav"}
      flexDirection={"column"}
      alignItems={"center"}
      gap={"36px"}
      h={"100vh"}
      p={"20px 16px"}
    >
      <Flex width={"40px"} height={"40px"} alignItems={"center"}>
        <Image src={"/logo.png"} />
      </Flex>
      <Flex flexDirection={"column"} gap={"16px"}>
        {children}
      </Flex>
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
