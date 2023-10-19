import { Flex, FlexProps } from "@chakra-ui/react";

export const Card = ({ children, ...flexProps }: FlexProps) => {
  return <Flex {...flexProps}>{children}</Flex>;
};
