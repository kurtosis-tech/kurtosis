import { Card, CardBody, CardHeader, Text } from "@chakra-ui/react";
import { PropsWithChildren, ReactElement } from "react";

type TitledCardProps = PropsWithChildren<{
  title: string;
  controls?: ReactElement;
}>;

export const TitledCard = ({ title, controls, children }: TitledCardProps) => {
  return (
    <Card variant={"titledCard"}>
      <CardHeader
        display={"flex"}
        justifyContent={"flex-start"}
        alignItems={"center"}
        width={"100%"}
        gap={"16px"}
        borderBottom={"1px solid"}
        borderBottomColor={"gray.500"}
      >
        <Text as={"span"} fontSize={"xs"} fontWeight={"semibold"}>
          {title}
        </Text>
        {controls}
      </CardHeader>
      <CardBody>{children}</CardBody>
    </Card>
  );
};
