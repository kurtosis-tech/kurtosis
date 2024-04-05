import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Icon, IconButton, Link } from "@chakra-ui/react";
import { FaLock, FaUnlock } from "react-icons/fa";
import { useEnclavesContext } from "../../EnclavesContext";
import { PortsTableRow } from "../tables/PortsTable";

type PortMaybeLinkProps = {
  port: PortsTableRow;
};

export const PortMaybeLink = ({ port }: PortMaybeLinkProps) => {
  const { lockUnlockPort } = useEnclavesContext();

  const isHttpLink = port.port.applicationProtocol?.startsWith("http");

  const handleLockUnlockClick = async () => {
    const lock = !port.port.locked;
    await lockUnlockPort(port.port.privatePort, port.serviceUuid, port.enclaveUuid, lock);
  };

  return (
    <>
      {isHttpLink ? (
        <Link href={port.link} isExternal>
          {port.port.name}&nbsp;
          <ExternalLinkIcon mx="2px" />
        </Link>
      ) : (
        port.port.name
      )}
      {port.port.locked !== undefined && (
        <IconButton
          aria-label={port.port.locked ? "Unlock port" : "Lock port"}
          icon={<Icon as={port.port.locked ? FaUnlock : FaLock} />}
          onClick={handleLockUnlockClick}
          ml={2}
          size="sm"
          colorScheme={port.port.locked ? "red" : "green"}
        />
      )}
    </>
  );
};
