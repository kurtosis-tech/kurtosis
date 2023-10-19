import { mode } from "@chakra-ui/theme-tools";
import { ChakraProvider, extendTheme, StyleFunctionProps, ThemeConfig } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import type { ChakraProviderProps } from "@chakra-ui/react/dist/chakra-provider";
import { tabsTheme } from "./theme/tabsTheme";
import Fonts from "./Fonts";

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
  colors: {
    kurtosis: {
      50: "#111111",
      100: "#1D1D1D",
      200: "#1E1E1E",
      300: "#2E2E2E",
      400: "#393B3E",
      700: "#00C223",
    },
  },
  styles: {
    global: (props: StyleFunctionProps) => ({
      "nav.primaryNav": {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "black")(props),
      },
      main: {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "kurtosis.50")(props),
        color: "gray.200",
      },
    }),
  },
  components: {
    Button: {
      variants: {
        nav: {
          _active: {
            bg: "kurtosis.300",
            color: "kurtosis.700",
          },
          color: "white",
          borderWidth: "1px",
          borderColor: "kurtosis.300",
        },
      },
    },
    Breadcrumb: {
      variants: {
        topNavigation: {
          link: {
            "&[aria-current=page]": {
              color: "whiteAlpha.700",
            },
          },
          separator: {
            color: "gray.100",
          },
        },
      },
    },
    Card: {
      baseStyle: {
        container: {
          bg: "kurtosis.200",
          borderWidth: "1px",
          borderColor: "kurtosis.400",
          borderRadius: "8px",
          padding: "1rem",
        },
      },
    },
    Tabs: tabsTheme,
  },
});

export const KurtosisThemeProvider = ({
  children,
  ...chakraProps
}: PropsWithChildren<Omit<ChakraProviderProps, "theme">>) => {
  return (
    <ChakraProvider theme={theme} {...chakraProps}>
      <Fonts />
      {children}
    </ChakraProvider>
  );
};
