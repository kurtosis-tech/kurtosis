import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Icon, Link, Text } from "@chakra-ui/react";
import { FaLock, FaUnlock } from "react-icons/fa";
import { PortsTableRow } from "../tables/PortsTable";

type PortMaybeLinkProps = {
  port: PortsTableRow;
};

export const PortMaybeLink = ({ port }: PortMaybeLinkProps) => {
  const isHttpLink = port.port.applicationProtocol?.startsWith("http");

  return (
    <Text>
      {isHttpLink ? (
        <Link href={port.link} isExternal>
          {port.port.name}&nbsp;
          <ExternalLinkIcon mx="2px" />
          {port.port.locked !== undefined && (
            <Icon as={port.port.locked ? FaLock : FaUnlock} ml={2} color={port.port.locked ? "red.500" : "green.500"} />
          )}
        </Link>
      ) : (
        port.port.name
      )}
    </Text>
  );
};
