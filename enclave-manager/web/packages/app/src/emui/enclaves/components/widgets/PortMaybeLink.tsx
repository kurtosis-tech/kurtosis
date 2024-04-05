import { ExternalLinkIcon, LockIcon, UnlockIcon } from "@chakra-ui/icons";
import { IconButton, Link, Text, Tooltip, useClipboard, useToast } from "@chakra-ui/react";
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
    <>
      {isHttpLink ? (
        <Link href={port.link} isExternal display="flex" alignItems="center">
          {port.port.locked !== undefined && (
            <Tooltip
              label={port.port.locked ? "Publish port" : "Make port private"}
              fontSize="small"
            >
              <IconButton
                icon={port.port.locked ? <LockIcon /> : <UnlockIcon />}
                variant={"ghost"}
                size={"xs"}
                cursor="pointer"
                onClick={() => handleLockUnlockClick}
                aria-label={port.port.locked ? "Publish port" : "Make port private"}
              />
            </Tooltip>
          )}
          <Text>{port.port.name}&nbsp;</Text>
          <ExternalLinkIcon ml="2px" />
        </Link>
      ) : (
        port.port.name
      )}
    </>
  );
};
