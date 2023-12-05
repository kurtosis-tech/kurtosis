import { tagAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, StyleFunctionProps } from "@chakra-ui/react";

const { defineMultiStyleConfig } = createMultiStyleConfigHelpers(tagAnatomy.keys);

// export the component theme
export const tagTheme = defineMultiStyleConfig({
  baseStyle: {
    container: { textTransform: "uppercase" },
  },
  variants: {
    asText: (props: StyleFunctionProps) => ({
      container: {
        bg: "none",
        color: `${props.colorScheme}.400`,
        padding: 0,
        fontSize: "inherit",
        lineHeight: "inherit",
        fontWeight: "semibold",
      },
    }),
    square: (props: StyleFunctionProps) => ({
      container: {
        bg: `${props.colorScheme}.900`,
        color: `${props.colorScheme}.400`,
        padding: "0 4px",
        fontSize: "xs",
        lineHeight: "16px",
        borderRadius: "2px",
        fontWeight: "bold",
        minHeight: "unset",
      },
    }),
    progress: (props: StyleFunctionProps) => ({
      container: {
        bg: `${props.colorScheme}.900`,
        color: `${props.colorScheme}.100`,
      },
    }),
    solid: (props: StyleFunctionProps) => ({
      container: {
        color: `${props.colorScheme}.400`,
        bg: "gray.700",
      },
    }),
  },
});
