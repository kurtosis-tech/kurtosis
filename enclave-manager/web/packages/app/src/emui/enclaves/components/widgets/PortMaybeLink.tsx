import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Icon, Link, Text, Tooltip, useClipboard, useToast } from "@chakra-ui/react";
import { FaLock, FaUnlock } from "react-icons/fa";
import { useEnclavesContext } from "../../EnclavesContext";
import { PortsTableRow } from "../tables/PortsTable";

type PortMaybeLinkProps = {
  port: PortsTableRow;
};

export const PortMaybeLink = ({ port }: PortMaybeLinkProps) => {
  const { lockUnlockPort } = useEnclavesContext();
  const toast = useToast();
  const { onCopy } = useClipboard(port.link);

  const isHttpLink = port.port.applicationProtocol?.startsWith("http");

  const handleLockUnlockClick = async (e: React.MouseEvent<SVGElement>) => {
    e.preventDefault();
    const lock = !port.port.locked;
    await lockUnlockPort(port.port.privatePort, port.port.serviceShortUuid, port.port.enclaveShortUuid, lock);
    onCopy();
    toast({
      title: `Link ${lock ? "locked" : "unlocked"}`,
      description: `The link has been ${lock ? "locked" : "unlocked"} and copied to the clipboard.`,
      status: lock ? "success" : "warning",
      duration: 3000,
      isClosable: true,
    });
  };

  return (
    <Text>
      {isHttpLink ? (
        <Link href={port.link} isExternal>
          {port.port.name}&nbsp;
          <ExternalLinkIcon mx="2px" />
          {port.port.locked !== undefined && (
              // add a Tooltip here that tells you what is going to happen
              // adding it with GPT made it stuck on the top left of the screen
              <Icon
                as={port.port.locked ? FaLock : FaUnlock}
                ml={2}
                color={port.port.locked ? "red.500" : "green.500"}
                cursor="pointer"
                size="sm"
                onClick={handleLockUnlockClick}
              />
          )}
        </Link>
      ) : (
        port.port.name
      )}
    </Text>
  );
};
