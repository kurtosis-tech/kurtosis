import { tabsAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers } from "@chakra-ui/react";

const { definePartsStyle, defineMultiStyleConfig } = createMultiStyleConfigHelpers(tabsAnatomy.keys);

// define the base component styles
const baseStyle = definePartsStyle({
  // define the part you're going to style
  tab: {
    color: "green",
    fontWeight: "semibold", // change the font weight
  },
  tabpanel: {
    fontFamily: "mono", // change the font family
  },
});

// export the component theme
export const tabsTheme = defineMultiStyleConfig({
  variants: {
    "soft-rounded": {
      tab: {
        fontStyle: "normal",
        fontWeight: 500,
        fontSize: "18px",
        color: "gray.200",
        lineHeight: "28px",
        _selected: {
          fontWeight: 600,
        },
      },
      tablist: {
        marginBottom: "1.5rem",
      },
    },
  },
});
