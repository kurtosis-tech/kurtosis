import { Card, CardBody, CardHeader, CardProps, Flex, Text } from "@chakra-ui/react";
import { PropsWithChildren, ReactElement } from "react";

type TitledCardProps = CardProps &
  PropsWithChildren<{
    title: string;
    controls?: ReactElement;
    rightControls?: ReactElement;
  }>;

export const TitledCard = ({ title, controls, rightControls, children, ...cardProps }: TitledCardProps) => {
  return (
    <Card variant={"titledCard"} {...cardProps}>
      <CardHeader
        display={"flex"}
        justifyContent={"space-between"}
        alignItems={"center"}
        width={"100%"}
        gap={"16px"}
        borderBottom={"1px solid"}
        borderBottomColor={"gray.500"}
      >
        <Flex justifyContent={"flex-start"} gap={"16px"} alignItems={"center"}>
          <Text as={"span"} fontSize={"xs"} fontWeight={"semibold"} wordBreak={"break-all"}>
            {title}
          </Text>
          {controls}
        </Flex>
        <Flex>{rightControls}</Flex>
      </CardHeader>
      <CardBody overflow={"auto"}>{children}</CardBody>
    </Card>
  );
};
