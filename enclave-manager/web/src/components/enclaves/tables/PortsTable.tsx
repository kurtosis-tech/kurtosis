import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Flex, Link, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { CopyButton } from "../../CopyButton";
import { DataTable } from "../../DataTable";
import { transportProtocolToString } from "../utils";

const columnHelper = createColumnHelper<Port>();

type PortsTableProps = {
  ports: Port[];
  ip: string;
  isPublic?: boolean;
};

export const PortsTable = ({ ports, ip, isPublic }: PortsTableProps) => {
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
            {isPublic && (
              <Link
                href={`${row.original.maybeApplicationProtocol}://${ip}:${row.original.number}`}
                target="_blank"
                rel="noopener noreferrer"
                isExternal
              >
                {row.original.maybeApplicationProtocol}://{ip}:{row.original.number} <ExternalLinkIcon mx="2px" />
              </Link>
            )}
            {!isPublic && `${row.original.maybeApplicationProtocol}://${ip}:${row.original.number}`}
          </Text>
        ),
      }),
      columnHelper.display({
        id: "copyButton",
        cell: ({ row }) => (
          <Flex justifyContent={"flex-end"}>
            <CopyButton
              contentName={"link"}
              isIconButton
              aria-label={"Copy this port"}
              valueToCopy={`${row.original.maybeApplicationProtocol}://${ip}:${row.original.number}`}
            />
          </Flex>
        ),
      }),
    ],
    [ip, isPublic],
  );

  return <DataTable columns={columns} data={ports} defaultSorting={[{ id: "number", desc: true }]} />;
};
