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
      tab: {
        fontStyle: "normal",
        fontWeight: 500,
        fontSize: "18px",
        color: "gray.200",
        lineHeight: "28px",
        _selected: {
          fontWeight: 600,
          color: `${props.colorScheme}.400`,
          bg: `gray.800`,
        },
      },
      tablist: {
        marginBottom: "8px",
      },
      tabpanel: {
        padding: "12px",
      },
    }),
  },
});
