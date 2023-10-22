import { mode } from "@chakra-ui/theme-tools";
import { ChakraProvider, extendTheme, StyleFunctionProps, ThemeConfig } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import type { ChakraProviderProps } from "@chakra-ui/react/dist/chakra-provider";
import { tabsTheme } from "./theme/tabsTheme";
import Fonts from "./theme/Fonts";
import { tagTheme } from "./theme/tagsTheme";

const config: ThemeConfig = {
  initialColorMode: "dark",
  useSystemColorMode: false,
  disableTransitionOnChange: false,
};

const theme = extendTheme({
  config,
  fonts: {
    heading: `'Inter', sans-serif`,
    body: `'Inter', sans-serif`,
  },
  colors: {
    kurtosisGreen: {
      100: "#005e11",
      200: "#008c19",
      300: "#00bb22",
      400: "#00C223", // The true green
      500: "#33ee55",
      600: "#66f27f",
      700: "#99f7aa",
    },
    kurtosisGray: {
      50: "#111111", // ui background
      100: "#1D1D1D", // selected background
      200: "#1E1E1E",
      300: "#2E2E2E",
      400: "#393B3E",
      500: "#5B5B5B", // icon color
      600: "#606770",
      700: "#878787",
      900: "#E3E3E3", // text
    },
  },
  styles: {
    global: (props: StyleFunctionProps) => ({
      "nav.primaryNav": {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "black")(props),
      },
      main: {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "kurtosisGray.50")(props),
        color: "kurtosisGray.900",
      },
    }),
  },
  components: {
    Button: {
      variants: {
        kurtosisOutline: (props: StyleFunctionProps) => {
          const outline = theme.components.Button.variants!.outline(props);
          return {
            ...outline,
            _hover: { ...outline._hover, bg: "initial", borderColor: `${props.colorScheme}.400` },
            color: `${props.colorScheme}.400`,
            borderColor: "kurtosisGray.600",
          };
        },
        kurtosisGroupOutline: (props: StyleFunctionProps) => {
          const outline = theme.components.Button.variants!.outline(props);
          return {
            ...outline,
            _hover: { ...outline._hover, bg: "kurtosisGray.200" },
            color: `${props.colorScheme}.400`,
            borderColor: "kurtosisGray.600",
          };
        },
        nav: {
          _active: {
            bg: "kurtosisGray.300",
            color: "kurtosisGreen.400",
          },
          _hover: {
            bg: "kurtosisGray.300",
            color: "white",
          },
          color: "kurtosisGray.700",
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
          bg: "kurtosisGray.200",
          borderWidth: "1px",
          borderColor: "kurtosisGray.400",
          borderRadius: "8px",
          padding: "1rem",
        },
      },
    },
    Table: {
      variants: {
        kurtosis: {
          th: {
            color: "kurtosisGray.900",
            borderBottom: "1px solid",
            borderColor: "kurtosisGray.400",
          },
        },
      },
    },
    Tabs: tabsTheme,
    Tag: tagTheme,
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
