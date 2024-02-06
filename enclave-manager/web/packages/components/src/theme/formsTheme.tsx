import { formAnatomy as parts } from "@chakra-ui/anatomy";
import { createMultiStyleConfigHelpers, cssVar, defineStyle } from "@chakra-ui/styled-system";

const { definePartsStyle, defineMultiStyleConfig } = createMultiStyleConfigHelpers(parts.keys);

const $fg = cssVar("form-control-color");

const baseStyleRequiredIndicator = defineStyle({
  marginStart: "1",
  [$fg.variable]: "colors.red.400",
  _dark: {
    [$fg.variable]: "colors.red.300",
  },
  color: $fg.reference,
});

const baseStyleHelperText = defineStyle({
  mt: "2",
  [$fg.variable]: "colors.gray.600",
  _dark: {
    [$fg.variable]: "colors.whiteAlpha.600",
  },
  color: "gray.100",
  lineHeight: "normal",
  fontSize: "xs",
});

const baseStyle = definePartsStyle({
  container: {
    width: "100%",
    position: "relative",
  },
  requiredIndicator: baseStyleRequiredIndicator,
  helperText: baseStyleHelperText,
});

export const formsTheme = defineMultiStyleConfig({
  baseStyle,
});
