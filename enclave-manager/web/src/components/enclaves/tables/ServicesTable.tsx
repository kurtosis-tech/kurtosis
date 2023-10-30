import { Button } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import {
  GetServicesResponse,
  Port,
  ServiceInfo,
  ServiceStatus,
} from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { RemoveFunctions } from "../../../utils/types";
import { DataTable } from "../../DataTable";
import { ImageButton } from "../widgets/ImageButton";
import { PortsSummary } from "../widgets/PortsSummary";
import { ServiceStatusTag } from "../widgets/ServiceStatus";

type ServicesTableRow = {
  serviceUUID: string;
  name: string;
  status: ServiceStatus;
  // started: DateTime | null; TODO: The api needs to support this field
  image?: string;
  ports: { privatePorts: Port[]; publicPorts: Port[] };
};

const serviceToRow = (service: ServiceInfo): ServicesTableRow => {
  return {
    serviceUUID: service.serviceUuid,
    name: service.name,
    status: service.serviceStatus,
    image: service.container?.imageName,
    ports: {
      privatePorts: Object.values(service.privatePorts),
      publicPorts: Object.values(service.maybePublicPorts),
    },
  };
};

const columnHelper = createColumnHelper<ServicesTableRow>();

type EnclavesTableProps = {
  enclaveShortUUID: string;
  servicesResponse: RemoveFunctions<GetServicesResponse>;
};

export const ServicesTable = ({ enclaveShortUUID, servicesResponse }: EnclavesTableProps) => {
  const services = Object.values(servicesResponse.serviceInfo).map(serviceToRow);

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
        cell: (statusCell) => <ServiceStatusTag status={statusCell.getValue()} />,
      }),
      columnHelper.accessor("image", {
        header: "Image",
        cell: (imageCell) => <ImageButton image={imageCell.getValue()} />,
      }),
      columnHelper.accessor("ports", {
        header: "Ports",
        cell: (portsCell) => (
          <PortsSummary
            privatePorts={portsCell.getValue().privatePorts}
            publicPorts={portsCell.getValue().publicPorts}
          />
        ),
        sortingFn: (a, b) =>
          a.original.ports.publicPorts.length +
          a.original.ports.privatePorts.length -
          b.original.ports.publicPorts.length -
          b.original.ports.privatePorts.length,
      }),
      columnHelper.accessor("serviceUUID", {
        header: "Logs",
        cell: (portsCell) => (
          <Link to={`/enclave/${enclaveShortUUID}/service/${portsCell.getValue()}/logs`}>
            <Button size={"xs"} variant={"ghost"}>
              View
            </Button>
          </Link>
        ),
        enableSorting: false,
      }),
    ],
    [],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
