import { Flex } from "@chakra-ui/react";
import { PropsWithChildren, useRef } from "react";
import { Navbar } from "../emui/Navbar";
import { KurtosisBreadcrumbs } from "./KurtosisBreadcrumbs";
import {
  MAIN_APP_BOTTOM_PADDING,
  MAIN_APP_LEFT_PADDING,
  MAIN_APP_MAX_WIDTH,
  MAIN_APP_RIGHT_PADDING,
  MAIN_APP_TOP_PADDING,
} from "./theme/constants";

export const AppLayout = ({ children }: PropsWithChildren) => {
  return (
    <>
      <Navbar />
      <Flex
        as="main"
        w={"100%"}
        minH={"100vh"}
        justifyContent={"flex-start"}
        flexDirection={"column"}
        className={"app-container"}
      >
        {children}
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
        flexDirection={"column"}
        w={"100%"}
        h={"100%"}
        maxHeight={preventPageScroll ? `100vh` : undefined}
        flex={"1"}
      >
        <Flex
          flexDirection={"column"}
          flex={"1"}
          w={"100%"}
          h={"100%"}
          maxWidth={MAIN_APP_MAX_WIDTH}
          pl={MAIN_APP_LEFT_PADDING}
          pr={MAIN_APP_RIGHT_PADDING}
        >
          <KurtosisBreadcrumbs />
          <Flex
            w={"100%"}
            h={"100%"}
            pt={MAIN_APP_TOP_PADDING}
            pb={MAIN_APP_BOTTOM_PADDING}
            flexDirection={"column"}
            flex={"1"}
          >
            {children}
          </Flex>
        </Flex>
      </Flex>
    );
  }

  // TS cannot infer that children is an array if numberOfChildren === 2
  if (numberOfChildren === 2 && Array.isArray(children)) {
    return (
      <Flex direction="column" width={"100%"} h={"100%"} flex={"1"}>
        <Flex ref={headerRef} width={"100%"} bg={"gray.850"}>
          <Flex
            flexDirection={"column"}
            width={"100%"}
            pl={MAIN_APP_LEFT_PADDING}
            pr={MAIN_APP_RIGHT_PADDING}
            maxW={MAIN_APP_MAX_WIDTH}
          >
            <KurtosisBreadcrumbs />
            {children[0]}
          </Flex>
        </Flex>
        <Flex
          maxWidth={MAIN_APP_MAX_WIDTH}
          pl={MAIN_APP_LEFT_PADDING}
          pr={MAIN_APP_RIGHT_PADDING}
          pt={MAIN_APP_TOP_PADDING}
          pb={MAIN_APP_BOTTOM_PADDING}
          w={"100%"}
          h={"100%"}
          flex={"1"}
          flexDirection={"column"}
          maxHeight={preventPageScroll ? `calc(100vh - ${headerRef.current?.offsetHeight || 0}px)` : undefined}
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
