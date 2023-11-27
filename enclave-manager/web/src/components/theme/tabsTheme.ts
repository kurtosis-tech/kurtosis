import { tabsAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, StyleFunctionProps } from "@chakra-ui/react";

const { defineMultiStyleConfig } = createMultiStyleConfigHelpers(tabsAnatomy.keys);

// export the component theme
export const tabsTheme = defineMultiStyleConfig({
  defaultProps: {
    variant: "line",
    colorScheme: "kurtosisGreen",
  },
  variants: {
    line: (props: StyleFunctionProps) => ({
      root: {
        display: "flex",
        flexDirection: "column",
        height: "100%",
        width: "100%",
        flex: "1",
      },
      tablist: {
        height: "47px",
        borderColor: "transparent",
      },
      tab: {
        fontWeight: "md",
        fontSize: "sm",
        color: "gray.100",
        lineHeight: "28px",
        padding: "4px 16px 2px 16px",
        _active: {
          bg: "none",
        },
        textTransform: "capitalize",
      },
      tabpanels: {
        display: "flex",
        flexDirection: "column",
        height: "100%",
        flex: "1",
      },
      tabpanel: {
        display: "flex",
        flexDirection: "column",
        flex: "1",
        padding: "32px 0px",
        height: "100%",
      },
    }),
  },
});
