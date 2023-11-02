import { ChakraProvider, defineStyle, extendTheme, StyleFunctionProps, ThemeConfig, Tooltip } from "@chakra-ui/react";
import type { ChakraProviderProps } from "@chakra-ui/react/dist/chakra-provider";
import { mode } from "@chakra-ui/theme-tools";
import { PropsWithChildren } from "react";
import Fonts from "./theme/Fonts";
import { formsTheme } from "./theme/formsTheme";
import { tabsTheme } from "./theme/tabsTheme";
import { tagTheme } from "./theme/tagsTheme";

// https://github.com/chakra-ui/chakra-ui/issues/3347
Tooltip.defaultProps = {
  hasArrow: true,
  openDelay: 500,
  size: "sm",
};

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
    kurtosisSelected: {
      100: "#292929",
    },
    kurtosisGreen: {
      100: "#005e11",
      200: "#008c19",
      300: "#00bb22",
      400: "#00C223", // The true green
      500: "#33ee55",
      600: "#66f27f",
      700: "#99f7aa",
    },
    darkBlue: {
      400: "#516A77",
    },
    gray: {
      100: "#E3E3E3", // text
      200: "#878787",
      250: "#7A7A7A",
      300: "#606770",
      400: "#5B5B5B", // icon color
      500: "#393B3E",
      600: "#2E2E2E",
      650: "#292929",
      700: "#1E1E1E",
      800: "#1D1D1D", // selected background
      900: "#111111", // ui background
    },
  },
  fontSizes: {
    xs: "12px",
    sm: "14px",
    md: "16px",
    lg: "18px",
    xl: "20px",
    ["2xl"]: "22px",
  },
  styles: {
    global: (props: StyleFunctionProps) => ({
      body: {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "gray.900")(props),
      },
      "nav.primaryNav": {
        bg: mode(props.theme.semanticTokens.colors["chakra-body-bg"]._light, "black")(props),
      },
      main: {
        color: "gray.100",
        fontSize: "sm",
      },
    }),
  },
  components: {
    Badge: {
      baseStyle: {
        textTransform: "none",
        color: "gray.100",
      },
    },
    Button: {
      defaultProps: {
        variant: "outline",
      },
      variants: {
        outline: (props: StyleFunctionProps) => ({
          _hover: { bg: "initial", borderColor: `${props.colorScheme}.400` },
          _active: { bg: "initial" },
          color: `${props.colorScheme}.400`,
          borderColor: "gray.300",
        }),
        kurtosisGroupOutline: (props: StyleFunctionProps) => {
          const outline = theme.components.Button.variants!.outline(props);
          return {
            ...outline,
            _hover: { ...outline._hover, bg: "gray.700" },
            color: `${props.colorScheme}.400`,
            borderColor: "gray.300",
          };
        },
        kurtosisDisabled: (props: StyleFunctionProps) => {
          const outline = theme.components.Button.variants!.outline(props);
          return {
            ...outline,
            _hover: { ...outline._hover, bg: "gray.700" },
            bg: "gray.700",
            color: `${props.colorScheme}.100`,
            borderColor: "gray.300",
          };
        },
        solid: defineStyle((props) => ({
          _hover: { bg: "gray.700" },
          _active: { bg: "gray.700" },
          color: `${props.colorScheme}.400`,
          bg: "gray.700",
        })),
        ghost: defineStyle((props) => ({
          _hover: { bg: "gray.650" },
          color: `gray.100`,
        })),
        nav: {
          _active: {
            bg: "gray.600",
            color: "kurtosisGreen.400",
          },
          _hover: {
            bg: "gray.600",
            color: "white",
          },
          color: "gray.200",
        },
      },
    },
    Breadcrumb: {
      variants: {
        topNavigation: {
          link: {
            "&[aria-current=page]": {
              color: "gray.250",
            },
            fontSize: "sm",
            lineHeight: "24px",
          },
          separator: {
            color: "gray.250",
          },
        },
      },
    },
    Card: {
      baseStyle: {
        container: {
          bg: "gray.800",
          borderRadius: "8px",
          padding: "16px",
        },
      },
    },
    Checkbox: {
      defaultProps: {
        size: "md",
      },
      baseStyle: defineStyle(({ colorScheme }) => ({
        control: {
          borderColor: `gray.400`,
          _checked: {
            bg: `${colorScheme}.500`,
            borderColor: `${colorScheme}.500`,
            color: `white`,
            _hover: {
              bg: `${colorScheme}.500`,
              borderColor: `${colorScheme}.500`,
            },
          },
          _indeterminate: {
            bg: `${colorScheme}.500`,
            borderColor: `${colorScheme}.500`,
            color: `white`,
          },
        },
      })),
    },
    Form: formsTheme,
    Menu: {
      baseStyle: {
        list: {
          minW: "unset",
        },
      },
    },
    Popover: {
      baseStyle: {
        content: {
          bg: "gray.500",
          p: "8px",
        },
      },
    },
    Switch: {
      baseStyle: defineStyle((props) => ({
        track: {
          _checked: {
            bg: `${props.colorScheme}.500`,
          },
        },
      })),
    },

    Table: {
      variants: {
        simple: {
          th: {
            color: "gray.100",
            borderBottom: "1px solid",
            borderColor: "gray.500",
            textTransform: "uppercase",
          },
        },
      },
    },
    Tabs: tabsTheme,
    Tag: tagTheme,
    Tooltip: {
      sizes: {
        xs: defineStyle({
          fontSize: "12px",
          py: "2px",
          px: "6px",
          maxW: "200px",
        }),
        sm: defineStyle({
          fontSize: "sm",
          py: "1",
          px: "2",
          maxW: "200px",
        }),
        md: defineStyle({
          fontSize: "md",
          py: "2",
          px: "3",
          maxW: "300px",
        }),
        lg: defineStyle({
          fontSize: "lg",
          py: "2",
          px: "4",
          maxW: "350px",
        }),
      },
      baseStyle: {
        bg: "gray.500",
        //https://github.com/chakra-ui/chakra-ui/issues/4695
        ["--popper-arrow-bg" as string]: "colors.gray.500",
        color: "gray.100",
      },
      defaultProps: {
        size: "xs",
      },
    },
  },
});

export const KurtosisThemeProvider = ({
  children,
  ...chakraProps
}: PropsWithChildren<Omit<ChakraProviderProps, "theme">>) => {
  return (
    <ChakraProvider theme={theme} toastOptions={{ defaultOptions: { position: "top" } }} {...chakraProps}>
      <Fonts />
      {children}
    </ChakraProvider>
  );
};
