import { Flex, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { CopyButton } from "../../CopyButton";
import { DataTable } from "../../DataTable";
import { transportProtocolToString } from "../utils";

type PortsTableRow = {
  port: { transportProtocol: string; privatePort: number; name: string };
  link: string;
};

const getPortTableRows = (privatePorts: Record<string, Port>, publicPorts: Record<string, Port>): PortsTableRow[] => {
  return Object.entries(privatePorts).map(([name, port]) => ({
    port: { transportProtocol: transportProtocolToString(port.transportProtocol), privatePort: port.number, name },
    link: "Coming soon",
  }));
};

const columnHelper = createColumnHelper<PortsTableRow>();

type PortsTableProps = {
  privatePorts: Record<string, Port>;
  publicPorts: Record<string, Port>;
};

export const PortsTable = ({ privatePorts, publicPorts }: PortsTableProps) => {
  const columns = useMemo<ColumnDef<PortsTableRow, any>[]>(
    () => [
      columnHelper.accessor("port", {
        header: "Port",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>{row.original.port.name || "Unknown protocol"}</Text>
            <Text fontSize={"xs"} color={"gray.400"} fontWeight={"semibold"}>
              {row.original.port.privatePort}/{row.original.port.transportProtocol}
            </Text>
          </Flex>
        ),
      }),
      columnHelper.accessor("link", {
        header: "Link",
        minSize: 800,
        cell: ({ row }) => <Text width={"100%"}>{row.original.link}</Text>,
      }),
      columnHelper.display({
        id: "copyButton",
        cell: ({ row }) => (
          <Flex justifyContent={"flex-end"}>
            <CopyButton
              contentName={"link"}
              isIconButton
              aria-label={"Copy this port"}
              valueToCopy={`${row.original.port.privatePort}`}
            />
          </Flex>
        ),
      }),
    ],
    [],
  );

  return (
    <DataTable
      columns={columns}
      data={getPortTableRows(privatePorts, publicPorts)}
      defaultSorting={[{ id: "number", desc: true }]}
    />
  );
};
