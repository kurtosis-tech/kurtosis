import { tabsAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, StyleFunctionProps } from "@chakra-ui/react";

const { defineMultiStyleConfig } = createMultiStyleConfigHelpers(tabsAnatomy.keys);

// export the component theme
export const tabsTheme = defineMultiStyleConfig({
  defaultProps: {
    variant: "soft-rounded",
    colorScheme: "kurtosisGreen",
  },
  variants: {
    "soft-rounded": (props: StyleFunctionProps) => ({
      root: {
        height: "100%",
      },
      tab: {
        fontStyle: "normal",
        fontWeight: "medium",
        fontSize: "lg",
        color: "gray.100",
        lineHeight: "28px",
        _hover: {
          bg: `gray.700`,
        },
        _selected: {
          fontWeight: "semibold",
          color: `${props.colorScheme}.400`,
          bg: `gray.800`,
        },
        textTransform: "capitalize",
      },
      tabpanels: {
        height: "100%",
      },
      tabpanel: {
        padding: "32px 0px",
        height: "100%",
      },
    }),
  },
});
