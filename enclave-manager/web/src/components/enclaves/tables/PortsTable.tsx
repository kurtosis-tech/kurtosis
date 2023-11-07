import { Link, Flex, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { DataTable } from "../../DataTable";
import { transportProtocolToString } from "../utils";
import { CopyButton } from "../../CopyButton";

const columnHelper = createColumnHelper<Port>();

type PortsTableProps = {
  ports: Port[];
  ip: string;
};

export const PortsTable = ({ ports, ip }: PortsTableProps) => {
  const columns = useMemo<ColumnDef<Port, any>[]>(
    () => [
      columnHelper.accessor("number", {
        header: "Port",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>{row.original.maybeApplicationProtocol || "Unknown protocol"}</Text>
            <Text fontSize={"xs"} color={"gray.400"} fontWeight={"semibold"}>
              {row.original.number}/{transportProtocolToString(row.original.transportProtocol)}
            </Text>
          </Flex>
        ),
      }),
      columnHelper.accessor("maybeApplicationProtocol", {
        header: "Link",
        minSize: 800,
        cell: ({ row }) => (
          <Text width={"100%"}>
            <Link
              href={`${row.original.maybeApplicationProtocol}://${ip}:${row.original.number}`}
              target="_blank"
              rel="noopener noreferrer"
              isExternal
            >
              {row.original.maybeApplicationProtocol}://{ip}:{row.original.number}
            </Link>
          </Text>
        ),
      }),
      columnHelper.display({
        id: "copyButton",
        cell: ({ row }) => (
          <Flex justifyContent={"flex-end"}>
            <CopyButton
              contentName={"link"}
              valueToCopy={`${row.original.maybeApplicationProtocol}://${ip}:${row.original.number}`}
            />
          </Flex>
        ),
      }),
    ],
    [],
  );

  return <DataTable columns={columns} data={ports} defaultSorting={[{ id: "number", desc: true }]} />;
};
