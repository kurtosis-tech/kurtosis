import { Button, Card, Flex, Text, useToast } from "@chakra-ui/react";
import { ReactElement } from "react";
import { FiCopy } from "react-icons/fi";
import { assertDefined, isDefined } from "../utils";

type ValueCardProps = {
  title: string;
  value: string | ReactElement;
  copyEnabled?: boolean;
  copyValue?: string;
};

export const ValueCard = ({ title, value, copyEnabled, copyValue }: ValueCardProps) => {
  const toast = useToast();

  const handleCopyClick = () => {
    const valueToUse = isDefined(copyValue) ? copyValue : typeof value === "string" ? value : null;
    assertDefined(valueToUse, `Cannot work out which value to copy in ${title} value card`);
    navigator.clipboard.writeText(valueToUse);
    toast({
      title: `Copied '${valueToUse}' to the clipboard`,
      status: `success`,
    });
  };

  return (
    <Card height={"100%"} display={"flex"} flexDirection={"column"} justifyContent={"space-between"} gap={"16px"}>
      <Flex flexDirection={"row"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
        <Text fontSize={"sm"} fontWeight={"extrabold"} textTransform={"uppercase"} color={"gray.400"}>
          {title}
        </Text>
        {copyEnabled && (
          <Button leftIcon={<FiCopy />} size={"xs"} colorScheme={"darkBlue"} onClick={handleCopyClick}>
            Copy
          </Button>
        )}
      </Flex>
      <Text fontSize={"xl"}>{value}</Text>
    </Card>
  );
};
