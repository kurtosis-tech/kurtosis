import { Text, TextProps } from "@chakra-ui/react";

export const FindCommand = (props: TextProps) => {
  let text = "Ctrl + F";

  if (navigator.userAgent.indexOf("Mac") > -1) {
    text = "⌘F";
  }

  return (
    <Text as={"span"} {...props}>
      {text}
    </Text>
  );
};

export const OmniboxCommand = (props: TextProps) => {
  let text = "Ctrl + K";

  if (navigator.userAgent.indexOf("Mac") > -1) {
    text = "⌘K";
  }

  return (
    <Text as={"span"} {...props}>
      {text}
    </Text>
  );
};
