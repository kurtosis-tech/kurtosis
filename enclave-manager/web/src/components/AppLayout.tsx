import { Box, Flex } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
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
      <Flex flexDirection="column" as="main" w={"100%"} minH={"100vh"} className={"app-container"}>
        {children}
      </Flex>
    </>
  );
};

type AppPageLayoutProps = PropsWithChildren<{
  preventPageScroll?: boolean;
}>;

export const AppPageLayout = ({ preventPageScroll, children }: AppPageLayoutProps) => {
  const numberOfChildren = Array.isArray(children) ? children.length : 1;

  if (numberOfChildren === 1) {
    return (
      <Box w={"100%"} h={preventPageScroll ? `100vh` : "100%"}>
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
          <Box
            w={"100%"}
            h={"100%"}
            minH={preventPageScroll ? "0" : undefined}
            pt={MAIN_APP_TOP_PADDING}
            pb={MAIN_APP_BOTTOM_PADDING}
            flexDirection={"column"}
            flex={"1"}
          >
            {children}
          </Box>
        </Flex>
      </Box>
    );
  }

  // TS cannot infer that children is an array if numberOfChildren === 2
  if (numberOfChildren === 2 && Array.isArray(children)) {
    return (
      <Flex flexDirection={"column"} width={"100%"} h={preventPageScroll ? `100vh` : "100%"} flex={"1"}>
        <Box width={"100%"} bg={"gray.850"}>
          <Box width={"100%"} pl={MAIN_APP_LEFT_PADDING} pr={MAIN_APP_RIGHT_PADDING} maxW={MAIN_APP_MAX_WIDTH}>
            <KurtosisBreadcrumbs />
            {children[0]}
          </Box>
        </Box>
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
          minH={preventPageScroll ? "0" : undefined}
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
