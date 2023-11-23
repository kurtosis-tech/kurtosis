import { Card, CardBody, CardHeader, CardProps, Text } from "@chakra-ui/react";
import { PropsWithChildren, ReactElement } from "react";

type TitledCardProps = CardProps &
  PropsWithChildren<{
    title: string;
    controls?: ReactElement;
  }>;

export const TitledCard = ({ title, controls, children, ...cardProps }: TitledCardProps) => {
  return (
    <Card variant={"titledCard"} {...cardProps}>
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
      <CardBody overflow={"auto"}>{children}</CardBody>
    </Card>
  );
};
