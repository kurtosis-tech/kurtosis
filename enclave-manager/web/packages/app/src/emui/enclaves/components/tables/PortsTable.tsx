import { ExternalLinkIcon } from "@chakra-ui/icons";
import { Flex, Link, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Port } from "enclave-manager-sdk/build/api_container_service_pb";
import { DataTable, isDefined } from "kurtosis-ui-components";
import { useMemo } from "react";
import { KURTOSIS_CLOUD_HOST, KURTOSIS_CLOUD_PROTOCOL } from "../../../../client/constants";
import { transportProtocolToString } from "../utils";

type PortsTableRow = {
  port: {
    transportProtocol: string;
    privatePort: number;
    publicPort: number;
    name: string;
    applicationProtocol: string;
  };
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
        name,
      },
      link: link,
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
  const columns = useMemo<ColumnDef<PortsTableRow, any>[]>(
    () => [
      columnHelper.accessor("port", {
        id: "port_name",
        header: "Name",
        cell: ({ row, getValue }) => (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text>
              {row.original.port.applicationProtocol?.startsWith("http") ? (
                <Link href={row.original.link} isExternal>
                  {row.original.port.name}&nbsp;
                  <ExternalLinkIcon mx="2px" />
                </Link>
              ) : (
                row.original.port.name
              )}
            </Text>
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
    ],
    [],
  );

  return (
    <DataTable
      columns={columns}
      data={getPortTableRows(instanceUUID, enclaveUUID, serviceUUID, privatePorts, publicPorts, publicIp)}
      defaultSorting={[{ id: "port_name", desc: true }]}
    />
  );
};
