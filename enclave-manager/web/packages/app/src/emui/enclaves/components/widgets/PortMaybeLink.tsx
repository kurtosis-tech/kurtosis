import { CheckCircleIcon, ExternalLinkIcon, LockIcon, UnlockIcon } from "@chakra-ui/icons";
import { Box, Flex, Heading, Icon, IconButton, Link, Text, Tooltip, useClipboard, useToast } from "@chakra-ui/react";
import { useEnclavesContext } from "../../EnclavesContext";
import { PortsTableRow } from "../tables/PortsTable";

type PortMaybeLinkProps = {
  port: PortsTableRow;
  disablePortLocking?: boolean;
};

export const PortMaybeLink = ({ port, disablePortLocking }: PortMaybeLinkProps) => {
  const { lockUnlockPort } = useEnclavesContext();
  const toast = useToast();
  const { onCopy } = useClipboard(port.link);

  const isHttpLink = port.port.applicationProtocol?.startsWith("http");

  const handleLockUnlockClick = async (e: React.MouseEvent<HTMLElement>) => {
    e.preventDefault();
    const lock = !port.port.locked;
    await lockUnlockPort(port.port.privatePort, port.port.serviceShortUuid, port.port.enclaveShortUuid, lock);
    onCopy();
    toast({
      status: lock ? "success" : "warning",
      duration: 3000,
      isClosable: true,
      position: "bottom-right",
      render: () => (
        <Flex color="white" p={3} bg={lock ? "green.500" : "yellow.500"} borderRadius={6} gap={4}>
          <Icon as={CheckCircleIcon} w={6} h={6} />
          <Box>
            <Heading as="h4" fontSize="md" fontWeight="500" color="white">
              Link {lock ? "locked" : "unlocked"}
            </Heading>
            <Text marginTop={1} color="white">
              The link has been {lock ? "locked" : "unlocked"} and copied to the clipboard.
            </Text>
          </Box>
        </Flex>
      ),
    });
  };

  return (
    <>
      {isHttpLink ? (
        <Flex alignItems="center">
          {port.port.locked !== undefined && (
            <Tooltip label={port.port.locked ? "Publish port" : "Make port private"} fontSize="small">
              <IconButton
                icon={port.port.locked ? <LockIcon /> : <UnlockIcon />}
                variant={"ghost"}
                size={"xs"}
                cursor="pointer"
                onClick={handleLockUnlockClick}
                pointerEvents={disablePortLocking ? "none" : "auto"}
                color={disablePortLocking ? "gray.200" : port.port.locked ? "red.500" : "green.500"}
                aria-label={port.port.locked ? "Publish port" : "Make port private"}
              />
            </Tooltip>
          )}
          <Link href={port.link} isExternal display="flex" alignItems="center">
            <Text>{port.port.name}&nbsp;</Text>
            <ExternalLinkIcon ml="2px" />
          </Link>
        </Flex>
      ) : (
        port.port.name
      )}
    </>
  );
};
