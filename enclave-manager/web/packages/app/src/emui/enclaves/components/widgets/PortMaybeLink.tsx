import { CheckCircleIcon, ExternalLinkIcon, LockIcon, UnlockIcon } from "@chakra-ui/icons";
import { Box, Flex, IconButton, Link, Heading, Text, Tooltip, useClipboard, useToast, Icon } from "@chakra-ui/react";
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
      status: "success",
      duration: 3000,
      isClosable: true,
      render: () => (
        <Flex
          color='white'
          p={3}
          bg='green.500'
          borderRadius={6}
          gap={4}
        >
          <Icon as={CheckCircleIcon} w={6} h={6} />
          <Box>
            <Heading
              as="h4"
              fontSize="md"
              fontWeight="500"
              color="white"
            >
              Link {lock ? "locked" : "unlocked"}
            </Heading>
            <Text marginTop={1} color="white">The link has been {lock ? "locked" : "unlocked"} and copied to the clipboard.</Text>
          </Box>
        </Flex>
      )
    });
  };

  return (
    <>
      {isHttpLink ? (
        <Link href={port.link} isExternal display="flex" alignItems="center">
          {port.port.locked === undefined && (
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
