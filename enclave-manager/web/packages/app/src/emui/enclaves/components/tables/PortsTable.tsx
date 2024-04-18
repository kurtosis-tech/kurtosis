import { Empty } from "@bufbuild/protobuf";
import { WarningIcon } from "@chakra-ui/icons";
import { Box, Flex, Heading, Icon, Input, Text, useToast, UseToastOptions } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { DataTable, isDefined } from "kurtosis-ui-components";
import { useMemo } from "react";
import { Result } from "true-myth";
import { KURTOSIS_CLOUD_HOST, KURTOSIS_CLOUD_PROTOCOL } from "../../../../client/constants";
import { instanceUUID } from "../../../../cookies";
import { useEnclavesContext } from "../../EnclavesContext";
import { transportProtocolToString } from "../utils";
import { PortMaybeLink } from "../widgets/PortMaybeLink";

export type PortsTableRow = {
  port: {
    transportProtocol: string;
    privatePort: number;
    publicPort: number;
    name: string;
    applicationProtocol: string;
    locked: boolean | undefined;
    enclaveShortUuid: string;
    serviceShortUuid: string;
    alias: string | undefined;
  };
  link: string;
};
const shortUUID = (fullUUID: string) => fullUUID.substring(0, 12);

export const getPortTableRows = (
  enclaveUUID: string,
  serviceUUID: string,
  privatePorts: Record<string, Port>,
  publicPorts: Record<string, Port>,
  publicIp: string,
  serviceName?: string,
): PortsTableRow[] => {
  return Object.entries(privatePorts).map(([name, port]) => {
    let link;
    if (isDefined(instanceUUID) && instanceUUID.length > 0) {
      link =
        `${KURTOSIS_CLOUD_PROTOCOL}://` +
        `${port.number}-${shortUUID(serviceUUID)}-${shortUUID(enclaveUUID)}-${shortUUID(instanceUUID)}` +
        `.${KURTOSIS_CLOUD_HOST}`;
    } else {
      link = `${port.maybeApplicationProtocol ? port.maybeApplicationProtocol + "://" : ""}${publicIp}:${
        publicPorts[name].number
      }`;
    }
    return {
      port: {
        applicationProtocol: port.maybeApplicationProtocol,
        transportProtocol: transportProtocolToString(port.transportProtocol),
        privatePort: port.number,
        publicPort: publicPorts[name].number,
        name: isDefined(serviceName) ? `${serviceName}:${name}` : name,
        locked: privatePorts[name].locked,
        enclaveShortUuid: shortUUID(enclaveUUID),
        serviceShortUuid: shortUUID(serviceUUID),
        alias: privatePorts[name].alias,
      },
      link: link,
    };
  });
};

const columnHelper = createColumnHelper<PortsTableRow>();

type PortsTableProps = {
  enclaveUUID: string;
  serviceUUID: string;
  privatePorts: Record<string, Port>;
  publicPorts: Record<string, Port>;
  publicIp: string;
};

const getPortAliasColumn = (
  toast: (options?: UseToastOptions) => void,
  privatePorts: Record<string, Port>,
  addAlias: (
    portNumber: number,
    serviceShortUUID: string,
    enclaveShortUUID: string,
    alias: string,
  ) => Promise<Result<Empty, string>>,
) => {
  if (!Object.values(privatePorts).some((port) => isDefined(port.alias))) {
    return [];
  }

  return [
    columnHelper.accessor("port", {
      id: "port_alias",
      header: "Alias",
      cell: ({ row, getValue }) => {
        const { alias, privatePort, serviceShortUuid, enclaveShortUuid } = row.original.port;
        const isAliasEmpty = !isDefined(alias) || alias === "";
        const isHttpLink = row.original.port.applicationProtocol?.startsWith("http");

        const handleAliasSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
          e.preventDefault();
          const inputAlias = e.currentTarget.elements.namedItem("alias") as HTMLInputElement;
          const newAlias = inputAlias.value.trim();
          console.log(`in handle submit alias ${newAlias} and ${inputAlias}`);
          if (isAliasEmpty && newAlias !== "") {
            const result: Result<Empty, string> = await addAlias(
              privatePort,
              serviceShortUuid,
              enclaveShortUuid,
              newAlias,
            );
            if (result.isErr) {
              console.error("Failed to add alias:", result.error);
              toast({
                status: "error",
                duration: 3000,
                isClosable: true,
                position: "bottom-right",
                render: () => (
                  <Flex color="white" p={3} bg="red.500" borderRadius={6} gap={4}>
                    <Icon as={WarningIcon} w={6} h={6} />
                    <Box>
                      <Heading as="h4" fontSize="md" fontWeight="500" color="white">
                        Couldn't add alias
                      </Heading>
                      <Text marginTop={1} color="white">
                        Perhaps that alias is taken; try again.
                      </Text>
                    </Box>
                  </Flex>
                ),
              });
            }
          }
        };

        return (
          <Flex flexDirection={"column"} gap={"10px"}>
            {isAliasEmpty && isHttpLink ? (
              <form onSubmit={handleAliasSubmit}>
                <Input name="alias" placeholder="Add alias" />
              </form>
            ) : (
              <Text>{alias}</Text>
            )}
          </Flex>
        );
      },
    }),
  ];
};

export const PortsTable = ({ enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp }: PortsTableProps) => {
  const { addAlias } = useEnclavesContext();
  const toast = useToast();

  const columns = useMemo<ColumnDef<PortsTableRow, any>[]>(
    () => [
      columnHelper.accessor("port", {
        id: "port_name",
        header: "Name",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <PortMaybeLink port={row.original} />
          </Flex>
        ),
      }),
      columnHelper.accessor("port", {
        id: "private_public_ports",
        header: "Private / Public Ports",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>
              {row.original.port.privatePort} / {row.original.port.publicPort}
            </Text>
          </Flex>
        ),
      }),
      columnHelper.accessor("port", {
        id: "port_protocol",
        header: "Application Protocol",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>{row.original.port.applicationProtocol}</Text>
          </Flex>
        ),
      }),
      columnHelper.accessor("port", {
        id: "port_transport",
        header: "Transport Protocol",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>{row.original.port.transportProtocol}</Text>
          </Flex>
        ),
      }),
      ...getPortAliasColumn(toast, privatePorts, addAlias),
    ],
    [addAlias, privatePorts],
  );

  return (
    <DataTable
      columns={columns}
      data={getPortTableRows(enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp)}
      defaultSorting={[{ id: "port_name", desc: true }]}
    />
  );
};
