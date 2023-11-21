import { Flex } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { Navbar } from "../emui/Navbar";
import { KurtosisBreadcrumbs } from "./KurtosisBreadcrumbs";

export const AppLayout = ({ children }: PropsWithChildren) => {
  return (
    <>
      <Navbar />
      <Flex as="main" w={"100%"} minH={"calc(100vh - 40px)"} justifyContent={"flex-start"} className={"app-container"}>
        <Flex direction={"column"} width={"100%"}>
          <KurtosisBreadcrumbs />
          {children}
        </Flex>
      </Flex>
    </>
  );
};
