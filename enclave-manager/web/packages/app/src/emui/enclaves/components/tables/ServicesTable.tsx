import { Button, Stack } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { GetServicesResponse, ServiceInfo, ServiceStatus } from "enclave-manager-sdk/build/api_container_service_pb";
import { DataTable, RemoveFunctions } from "kurtosis-ui-components";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { ImageButton } from "../widgets/ImageButton";
import { PortsSummary } from "../widgets/PortsSummary";
import { ServiceStatusTag } from "../widgets/ServiceStatus";
import { getPortTableRows, PortsTableRow } from "./PortsTable";
import TerminalAccessModal from "../modals/TerminalAccessModal"

type ServicesTableRow = {
  serviceUUID: string;
  name: string;
  status: ServiceStatus;
  // started: DateTime | null; TODO: The api needs to support this field
  image?: string;
  ports: PortsTableRow[];
};

const serviceToRow = (enclaveUUID: string, service: ServiceInfo): ServicesTableRow => {
  return {
    serviceUUID: service.shortenedUuid,
    name: service.name,
    status: service.serviceStatus,
    image: service.container?.imageName,
    ports: getPortTableRows(
      enclaveUUID,
      service.serviceUuid,
      service.privatePorts,
      service.maybePublicPorts,
      service.maybePublicIpAddr,
    ),
  };
};

const columnHelper = createColumnHelper<ServicesTableRow>();

type ServicesTableProps = {
  enclaveUUID: string;
  enclaveShortUUID: string;
  servicesResponse: RemoveFunctions<GetServicesResponse>;
};

export const ServicesTable = ({ enclaveUUID, enclaveShortUUID, servicesResponse }: ServicesTableProps) => {
  const services = Object.values(servicesResponse.serviceInfo).map((service) => serviceToRow(enclaveUUID, service));

  const columns = useMemo<ColumnDef<ServicesTableRow, any>[]>(
    () => [
      columnHelper.accessor("name", {
        header: "Name",
        cell: ({ row, getValue }) => (
          <Link to={`/enclave/${enclaveShortUUID}/service/${row.original.serviceUUID}`}>
            <Button size={"sm"} variant={"ghost"}>
              {getValue()}
            </Button>
          </Link>
        ),
      }),
      columnHelper.accessor("status", {
        header: "Status",
        cell: (statusCell) => <ServiceStatusTag status={statusCell.getValue()} variant={"square"} />,
      }),
      columnHelper.accessor("image", {
        header: "Image",
        cell: (imageCell) => <ImageButton image={imageCell.getValue()} />,
      }),
      columnHelper.accessor("ports", {
        header: "Endpoints",
        cell: (portsCell) => <PortsSummary ports={portsCell.getValue()} />,
        meta: { centerAligned: true },
      }),
      columnHelper.accessor("serviceUUID", {
        header: "Actions",
        cell: (portsCell) => (
          <Stack direction="row" spacing={4}>
            <TerminalAccessModal />
            <Link to={`/enclave/${enclaveShortUUID}/service/${portsCell.getValue()}/logs`}>
              <Button size={"xs"} colorScheme='green' variant={"outline"}>
                View Logs
              </Button>
            </Link>
          </Stack>
        ),
        enableSorting: false,
      }),
    ],
    [enclaveShortUUID],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
