import { Flex } from "@chakra-ui/react";
import React, { PropsWithChildren } from "react";

type AppLayoutProps = PropsWithChildren<{
  Nav: React.ReactElement;
}>;

export const AppLayout = ({ Nav, children }: AppLayoutProps) => {
  return (
    <Flex flexDirection={"row"}>
      {Nav}
      <Flex as="main" w={"100%"} p={"3rem 3rem 3rem 3rem"} className={"app-container"}>
        {children}
      </Flex>
    </Flex>
  );
};
