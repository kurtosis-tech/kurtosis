import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Link, Text } from "@chakra-ui/react";
import { PortsTableRow } from "../tables/PortsTable";
import { Icon } from "@chakra-ui/react";
import { FaLock, FaUnlock } from "react-icons/fa";

type PortMaybeLinkProps = {
  port: PortsTableRow;
};

export const PortMaybeLink = ({ port }: PortMaybeLinkProps) => {
  return (
    <Text>
      {port.port.applicationProtocol?.startsWith("http") ? (
        <Link href={port.link} isExternal>
          {port.port.name}&nbsp;
          <ExternalLinkIcon mx="2px" />
        </Link>
      ) : (
        port.port.name
      )}
      {port.port.locked !== undefined && (
        <Icon
          as={port.port.locked ? FaLock : FaUnlock}
          ml={2}
          color={port.port.locked ? "red.500" : "green.500"}
        />
      )}
    </Text>
  );
};