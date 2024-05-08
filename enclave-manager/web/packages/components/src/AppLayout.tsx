import { Box, Flex } from "@chakra-ui/react";
import { createContext, PropsWithChildren, ReactElement, useContext } from "react";
import { KurtosisBreadcrumbs } from "./KurtosisBreadcrumbs";
import {
  MAIN_APP_BOTTOM_PADDING,
  MAIN_APP_LEFT_PADDING_WITHOUT_NAV,
  MAIN_APP_LEFT_PADDING_WITH_NAV,
  MAIN_APP_MAX_WIDTH,
  MAIN_APP_RIGHT_PADDING,
  MAIN_APP_TOP_PADDING,
} from "./theme/constants";
import { isDefined } from "./utils";

type AppLayoutContextState = {
  hasNavbar: boolean;
};

const AppLayoutContext = createContext<AppLayoutContextState>({ hasNavbar: true });

type AppLayoutProps = PropsWithChildren<{
  navbar?: ReactElement;
}>;

export const AppLayout = ({ children, navbar }: AppLayoutProps) => {
  return (
    <>
      {isDefined(navbar) && navbar}
      <Flex flexDirection="column" as="main" w={"100%"} minH={"100vh"} className={"app-container"}>
        <AppLayoutContext.Provider value={{ hasNavbar: isDefined(navbar) }}>{children}</AppLayoutContext.Provider>
      </Flex>
    </>
  );
};

type AppPageLayoutProps = PropsWithChildren<{
  preventPageScroll?: boolean;
}>;

export const AppPageLayout = ({ preventPageScroll, children }: AppPageLayoutProps) => {
  const { hasNavbar } = useContext(AppLayoutContext);
  const numberOfChildren = Array.isArray(children) ? children.length : 1;

  if (numberOfChildren === 1) {
    return (
      <Flex
        w={"100%"}
        h={preventPageScroll ? `100vh` : "100%"}
        flex={"1"}
        justifyContent={"center"}
        alignItems={"center"}
      >
        <Flex
          position={"absolute"}
          top={"0"}
          bottom={"0"}
          flexDirection={"column"}
          w={"100%"}
          h={"100%"}
          minH={"100%"}
          maxWidth={MAIN_APP_MAX_WIDTH}
          pl={hasNavbar ? MAIN_APP_LEFT_PADDING_WITH_NAV : MAIN_APP_LEFT_PADDING_WITHOUT_NAV}
          pr={MAIN_APP_RIGHT_PADDING}
        >
          <KurtosisBreadcrumbs />
          <Flex
            w={"100%"}
            h={"100%"}
            minH={preventPageScroll ? "0" : undefined}
            pt={MAIN_APP_TOP_PADDING}
            pb={MAIN_APP_BOTTOM_PADDING}
            flexDirection={"column"}
            flex={"1"}
            gap={"16px"}
          >
            {children}
          </Flex>
        </Flex>
      </Flex>
    );
  }

  // TS cannot infer that children is an array if numberOfChildren === 2
  if (numberOfChildren === 3 && Array.isArray(children)) {
    return (
      <Flex flexDirection={"column"} width={"100%"} h={preventPageScroll ? `100vh` : "100%"} flex={"1"}>
        <Flex width={"100%"} bg={"gray.850"} justifyContent={"center"}>
          <Box
            width={"100%"}
            pl={hasNavbar ? MAIN_APP_LEFT_PADDING_WITH_NAV : MAIN_APP_LEFT_PADDING_WITHOUT_NAV}
            pr={MAIN_APP_RIGHT_PADDING}
            maxW={MAIN_APP_MAX_WIDTH}
          >
            <KurtosisBreadcrumbs />
            {children[0]}
          </Box>
        </Flex>
        <Flex flexDirection={"column"} alignItems={"left"} h={"20px"}>
          <Box
            width={"100%"}
            pt={"5px"}
            pl={hasNavbar ? MAIN_APP_LEFT_PADDING_WITH_NAV : MAIN_APP_LEFT_PADDING_WITHOUT_NAV}
            pr={MAIN_APP_RIGHT_PADDING}
            maxW={MAIN_APP_MAX_WIDTH}
          >
            {children[1]}
          </Box>
        </Flex>
        <Flex h={"100%"} flex={"1"} flexDirection={"column"} alignItems={"center"}>
          <Flex
            maxWidth={MAIN_APP_MAX_WIDTH}
            pl={hasNavbar ? MAIN_APP_LEFT_PADDING_WITH_NAV : MAIN_APP_LEFT_PADDING_WITHOUT_NAV}
            pr={MAIN_APP_RIGHT_PADDING}
            pt={MAIN_APP_TOP_PADDING}
            pb={MAIN_APP_BOTTOM_PADDING}
            w={"100%"}
            flexDirection={"column"}
            minH={preventPageScroll ? "0" : undefined}
            flex={"1"}
          >
            {children[2]}
          </Flex>
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
