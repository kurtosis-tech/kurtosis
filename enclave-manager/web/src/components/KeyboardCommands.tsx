import { Text, TextProps } from "@chakra-ui/react";

export const FindCommand = (props: TextProps) => {
  let text = "^F";

  if (navigator.userAgent.indexOf("Mac") > -1) {
    text = "⌘F";
  }

  return (
    <Text as={"span"} {...props}>
      {text}
    </Text>
  );
};
