import { Card, Flex, Text } from "@chakra-ui/react";
import { ReactElement } from "react";
import { isDefined } from "../utils";
import { CopyButton } from "./CopyButton";

type ValueCardProps = {
  title: string;
  value: string | ReactElement;
  copyEnabled?: boolean;
  copyValue?: string;
};

export const ValueCard = ({ title, value, copyEnabled, copyValue }: ValueCardProps) => {
  return (
    <Card height={"100%"} display={"flex"} flexDirection={"column"} justifyContent={"space-between"} gap={"16px"}>
      <Flex flexDirection={"row"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
        <Text fontSize={"sm"} fontWeight={"extrabold"} textTransform={"uppercase"} color={"gray.400"}>
          {title}
        </Text>
        {copyEnabled && (
          <CopyButton
            valueToCopy={isDefined(copyValue) ? copyValue : typeof value === "string" ? value : null}
            contentName={title}
          />
        )}
      </Flex>
      <Text fontSize={"xl"}>{value}</Text>
    </Card>
  );
};
