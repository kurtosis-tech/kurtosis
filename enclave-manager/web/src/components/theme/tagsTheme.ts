import { tagAnatomy } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, StyleFunctionProps } from "@chakra-ui/react";

const { defineMultiStyleConfig } = createMultiStyleConfigHelpers(tagAnatomy.keys);

// export the component theme
export const tagTheme = defineMultiStyleConfig({
  variants: {
    kurtosisSubtle: (props: StyleFunctionProps) => {
      return {
        container: {
          bg: `${props.colorScheme}.900`,
          color: `${props.colorScheme}.400`,
          padding: "0 4px",
          fontSize: "12px",
          lineHeight: "16px",
          borderRadius: "2px",
          textTransform: "uppercase",
          fontWeight: "700",
          minHeight: "unset",
        },
      };
    },
  },
});
