import { ChakraProvider, extendTheme, ThemeConfig } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { ChakraProviderProps } from "@chakra-ui/react/dist/chakra-provider";

const config: ThemeConfig = {
  initialColorMode: "dark",
  useSystemColorMode: false,
  disableTransitionOnChange: false,
};

const theme = extendTheme({
  config,
  fonts: {
    heading: `'Gilroy', sans-serif`,
    body: `'Gilroy', sans-serif`,
  },
  components: {
    Button: {
      baseStyle: {
        ghost: {
          bg: "#00C224FF",
        },
      },
    },
  },
});

export const KurtosisThemeProvider = ({
  children,
  ...chakraProps
}: PropsWithChildren<Omit<ChakraProviderProps, "theme">>) => {
  return (
    <ChakraProvider theme={theme} {...chakraProps}>
      {children}
    </ChakraProvider>
  );
};
