import { Flex, Text } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

type TitledBoxProps = PropsWithChildren<{
  title: string;
}>;

export const TitledBox = ({ title, children }: TitledBoxProps) => {
  return (
    <Flex flexDirection={"column"} alignItems={"center"} gap={"16px"}>
      <Flex justifyContent={"flex-start"} width={"100%"}>
        <Text fontSize={"lg"} fontWeight={"medium"}>
          {title}
        </Text>
      </Flex>
      {children}
    </Flex>
  );
};
