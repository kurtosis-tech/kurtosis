import { Card, Flex, Text } from "@chakra-ui/react";
import { PropsWithChildren } from "react";

type TitledCardProps = PropsWithChildren<{
  title: string;
}>;

export const TitledCard = ({ title, children }: TitledCardProps) => {
  return (
    <Card display={"flex"} flexDirection={"column"} alignItems={"center"} gap={"16px"}>
      <Flex justifyContent={"center"}>
        <Text fontSize={"md"} fontWeight={"semibold"}>
          {title}
        </Text>
      </Flex>
      {children}
    </Card>
  );
};
