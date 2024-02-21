import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Link, Text } from "@chakra-ui/react";
import { PortsTableRow } from "../tables/PortsTable";

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
    </Text>
  );
};
