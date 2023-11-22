import { Flex } from "@chakra-ui/react";
import React, { PropsWithChildren } from "react";
import { KurtosisBreadcrumbs } from "./KurtosisBreadcrumbs";
import { MAIN_APP_MAX_WIDTH } from "./theme/constants";

type AppLayoutProps = PropsWithChildren<{
  Nav: React.ReactElement;
}>;

export const AppLayout = ({ Nav, children }: AppLayoutProps) => {
  return (
    <>
      {Nav}
      <Flex
        as="main"
        w={"100%"}
        minH={"calc(100vh - 40px)"}
        justifyContent={"flex-start"}
        p={"20px 40px 20px 112px"}
        className={"app-container"}
      >
        <Flex maxWidth={MAIN_APP_MAX_WIDTH} w={"100%"}>
          <Flex direction={"column"} gap={"36px"} width={"100%"}>
            <KurtosisBreadcrumbs />
            {children}
          </Flex>
        </Flex>
      </Flex>
    </>
  );
};
