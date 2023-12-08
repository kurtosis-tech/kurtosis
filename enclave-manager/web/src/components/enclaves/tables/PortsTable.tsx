import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Flex, Icon, Link, Text, Tooltip } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { FiAlertTriangle } from "react-icons/fi";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { isDefined } from "../../../utils";
import { CopyButton } from "../../CopyButton";
import { DataTable } from "../../DataTable";
import { transportProtocolToString } from "../utils";

type PortsTableRow = {
  port: { transportProtocol: string; privatePort: number; name: string };
  link: string;
};
const shortUUID = (fullUUID: string) => fullUUID.substring(0, 12);

const getPortTableRows = (
  instanceUUID: string,
  enclaveUUID: string,
  serviceUUID: string,
  privatePorts: Record<string, Port>,
  publicPorts: Record<string, Port>,
  publicIp: string,
): PortsTableRow[] => {
  return Object.entries(privatePorts).map(([name, port]) => {
    let thisLine = "";
    if (isDefined(instanceUUID) && instanceUUID.length > 0) {
      thisLine = `https://${port.number}-${shortUUID(serviceUUID)}-${shortUUID(enclaveUUID)}-${shortUUID(
        instanceUUID,
      )}.cloud.kurtosis.com`;
    } else {
      thisLine = `${port.maybeApplicationProtocol ? port.maybeApplicationProtocol + "://" : ""}${publicIp}:${
        publicPorts[name].number
      }`;
    }
    return {
      port: {
        transportProtocol: transportProtocolToString(port.transportProtocol),
        privatePort: port.number,
        name,
      },
      link: thisLine,
    };
  });
};

const columnHelper = createColumnHelper<PortsTableRow>();

type PortsTableProps = {
  instanceUUID: string;
  enclaveUUID: string;
  serviceUUID: string;
  privatePorts: Record<string, Port>;
  publicPorts: Record<string, Port>;
  publicIp: string;
};

export const PortsTable = ({
  instanceUUID,
  enclaveUUID,
  serviceUUID,
  privatePorts,
  publicPorts,
  publicIp,
}: PortsTableProps) => {
  const kurtosisClient = useKurtosisClient();

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
        cell: ({ row }) => (
          <Text width={"100%"}>
            {row.original.link.startsWith("http") ? (
              <Link href={row.original.link} isExternal>
                {row.original.link}
                <ExternalLinkIcon mx="2px" />
              </Link>
            ) : (
              row.original.link
            )}
            {kurtosisClient.isRunningInCloud() && (
              <Tooltip
                label={
                  "Only enclaves started using the CLI will have their ports available. This port may not work if it was started using the app."
                }
                shouldWrapChildren
              >
                <Icon m="0 10px" as={FiAlertTriangle} color={"orange.400"} />
              </Tooltip>
            )}
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
              valueToCopy={`${row.original.link}`}
            />
          </Flex>
        ),
      }),
    ],
    [kurtosisClient],
  );

  return (
    <DataTable
      columns={columns}
      data={getPortTableRows(instanceUUID, enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp)}
      defaultSorting={[{ id: "port", desc: true }]}
    />
  );
};
