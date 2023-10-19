import { Flex } from "@chakra-ui/react";
import React, { PropsWithChildren } from "react";

type AppLayoutProps = PropsWithChildren<{
  Nav: React.ReactElement;
}>;

export const AppLayout = ({ Nav, children }: AppLayoutProps) => {
  return (
    <Flex flexDirection={"row"}>
      {Nav}
      <Flex w={"100%"} p={"1rem"}>
        {children}
      </Flex>
    </Flex>
  );
};
