import { Empty } from "@bufbuild/protobuf";
import { Flex, Input, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { DataTable, isDefined } from "kurtosis-ui-components";
import { useMemo, useState } from "react";
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

export const PortsTable = ({ enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp }: PortsTableProps) => {
  const { addAlias } = useEnclavesContext();
  const [editedAlias, setEditedAlias] = useState<string>("");

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
      ...getPortAliasColumn(privatePorts, addAlias, editedAlias, setEditedAlias),
    ],
    [addAlias, editedAlias],
  );

  return (
    <DataTable
      columns={columns}
      data={getPortTableRows(enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp)}
      defaultSorting={[{ id: "port_name", desc: true }]}
    />
  );
};

const getPortAliasColumn = (
  privatePorts: Record<string, Port>,
  addAlias: (
    portNumber: number,
    serviceShortUUID: string,
    enclaveShortUUID: string,
    alias: string,
  ) => Promise<Result<Empty, string>>,
  editedAlias: string,
  setEditedAlias: (alias: string) => void,
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

        const handleAliasChange = (e: React.ChangeEvent<HTMLInputElement>) => {
          setEditedAlias(e.target.value);
        };

        const handleAliasBlur = async () => {
          if (isAliasEmpty && editedAlias !== "") {
            const result: Result<Empty, string> = await addAlias(
              privatePort,
              serviceShortUuid,
              enclaveShortUuid,
              editedAlias,
            );
            if (result.isErr) {
              console.error("Failed to add alias:", result.error);
            }
          }
        };

        return (
          <Flex flexDirection={"column"} gap={"10px"}>
            {isAliasEmpty ? (
              <Input
                value={editedAlias}
                onChange={handleAliasChange}
                onBlur={handleAliasBlur}
                placeholder="Add alias"
              />
            ) : (
              <Text>{alias}</Text>
            )}
          </Flex>
        );
      },
    }),
  ];
};
