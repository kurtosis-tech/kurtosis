import { Flex } from "@chakra-ui/react";
import { PropsWithChildren, useRef } from "react";
import { Navbar } from "../emui/Navbar";
import { KurtosisBreadcrumbs } from "./KurtosisBreadcrumbs";
import {
  BREADCRUMBS_HEIGHT,
  MAIN_APP_LEFT_PADDING,
  MAIN_APP_MAX_WIDTH,
  MAIN_APP_PADDING,
  MAIN_APP_RIGHT_PADDING,
  MAIN_APP_TABPANEL_PADDING,
} from "./theme/constants";

export const AppLayout = ({ children }: PropsWithChildren) => {
  return (
    <>
      <Navbar />
      <Flex as="main" w={"100%"} minH={"100vh"} justifyContent={"flex-start"} className={"app-container"}>
        <Flex direction={"column"} width={"100%"}>
          <KurtosisBreadcrumbs />
          {children}
        </Flex>
      </Flex>
    </>
  );
};

type AppPageLayoutProps = PropsWithChildren<{
  preventPageScroll?: boolean;
}>;

export const AppPageLayout = ({ preventPageScroll, children }: AppPageLayoutProps) => {
  const headerRef = useRef<HTMLDivElement>(null);
  const numberOfChildren = Array.isArray(children) ? children.length : 1;

  if (numberOfChildren === 1) {
    return (
      <Flex
        maxWidth={MAIN_APP_MAX_WIDTH}
        p={MAIN_APP_PADDING}
        w={"100%"}
        h={"100%"}
        maxHeight={preventPageScroll ? `calc(100vh - ${BREADCRUMBS_HEIGHT})` : undefined}
      >
        {children}
      </Flex>
    );
  }

  // TS cannot infer that children is an array if numberOfChildren === 2
  if (numberOfChildren === 2 && Array.isArray(children)) {
    return (
      <Flex direction="column" width={"100%"} h={"100%"}>
        <Flex ref={headerRef} width={"100%"} bg={"gray.850"} pl={MAIN_APP_LEFT_PADDING} pr={MAIN_APP_RIGHT_PADDING}>
          {children[0]}
        </Flex>
        <Flex
          maxWidth={MAIN_APP_MAX_WIDTH}
          p={MAIN_APP_TABPANEL_PADDING}
          w={"100%"}
          h={"100%"}
          maxHeight={
            preventPageScroll
              ? `calc(100vh - ${BREADCRUMBS_HEIGHT} - ${headerRef.current?.offsetHeight || 0}px)`
              : undefined
          }
        >
          {children[1]}
        </Flex>
      </Flex>
    );
  }

  throw new Error(
    `AppPageLayout expects to receive exactly one or two children. ` +
      `If there are two children, the first child is the header section and the next child is the body. ` +
      `Otherwise the only child is the body.`,
  );
};
