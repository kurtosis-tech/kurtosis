import { Card, CardBody, CardHeader, Text } from "@chakra-ui/react";
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
    <Card variant={"valueCard"} height={"100%"}>
      <CardHeader>
        <Text fontSize={"sm"} fontWeight={"extrabold"} textTransform={"uppercase"} color={"gray.400"}>
          {title}
        </Text>
        {copyEnabled && (
          <CopyButton
            valueToCopy={isDefined(copyValue) ? copyValue : typeof value === "string" ? value : null}
            contentName={title}
          />
        )}
      </CardHeader>
      <CardBody>
        <Text as={"div"} fontSize={"xl"}>
          {value}
        </Text>
      </CardBody>
    </Card>
  );
};
