import { Flex } from "@chakra-ui/react";
import React, { PropsWithChildren } from "react";

type AppLayoutProps = PropsWithChildren<{
  Nav: React.ReactElement;
}>;

export const AppLayout = ({ Nav, children }: AppLayoutProps) => {
  return (
    <Flex flexDirection={"row"}>
      {Nav}
      <Flex as="main" w={"100%"} justifyContent={"flex-start"} p={"20px 40px"} className={"app-container"}>
        <Flex maxWidth={"1320px"} w={"100%"}>
          {children}
        </Flex>
      </Flex>
    </Flex>
  );
};
