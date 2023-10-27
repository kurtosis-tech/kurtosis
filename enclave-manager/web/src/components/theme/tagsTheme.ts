import { tagAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, StyleFunctionProps } from "@chakra-ui/react";

const { defineMultiStyleConfig } = createMultiStyleConfigHelpers(tagAnatomy.keys);

// export the component theme
export const tagTheme = defineMultiStyleConfig({
  baseStyle: (props: StyleFunctionProps) => ({
    container: {
      bg: `${props.colorScheme}.900`,
      color: `${props.colorScheme}.400`,
      padding: "0 4px",
      fontSize: "xs",
      lineHeight: "16px",
      borderRadius: "2px",
      textTransform: "uppercase",
      fontWeight: "bold",
      minHeight: "unset",
    },
  }),
  variants: {
    asText: (props: StyleFunctionProps) => ({
      container: {
        bg: "none",
        padding: 0,
        fontSize: "inherit",
        lineHeight: "inherit",
        fontWeight: "semibold",
      },
    }),
  },
});
