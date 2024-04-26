import {Box, Button, Flex, Heading, Icon, Input, Text, useToast, UseToastOptions} from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import {
    GetServicesResponse,
    ServiceInfo,
    ServiceStatus, StarlarkRunResponseLine
} from "enclave-manager-sdk/build/api_container_service_pb";
import {DataTable, isDefined, RemoveFunctions, stringifyError} from "kurtosis-ui-components";
import {useMemo, useState} from "react";
import {Link, NavigateFunction, useNavigate} from "react-router-dom";
import { ImageButton } from "../widgets/ImageButton";
import { PortsSummary } from "../widgets/PortsSummary";
import { ServiceStatusTag } from "../widgets/ServiceStatus";
import { getPortTableRows, PortsTableRow } from "./PortsTable";
import {useEnclavesContext} from "../../EnclavesContext";
import {EnclaveInfo} from "enclave-manager-sdk/build/engine_service_pb";
import {EnclaveFullInfo} from "../../types";

type ServicesTableRow = {
  serviceUUID: string;
  name: string;
  status: ServiceStatus;
  // started: DateTime | null; TODO: The api needs to support this field
  image?: string;
  ports: PortsTableRow[];
  setimage?: string
};

const serviceToRow = (enclaveUUID: string, service: ServiceInfo): ServicesTableRow => {
  return {
    serviceUUID: service.shortenedUuid,
    name: service.name,
    status: service.serviceStatus,
    image: service.container?.imageName,
    setimage: service.container?.imageName, // set to same as container image, initially
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
  enclave?: RemoveFunctions<EnclaveFullInfo>;
};

export const ServicesTable = ({ enclaveUUID, enclaveShortUUID, servicesResponse, enclave }: ServicesTableProps) => {
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
        cell: (imageCell) => <ImageButton image={imageCell.getValue()} serviceName={"postgres"} enclave={enclave} />,
      }),
      columnHelper.accessor("ports", {
        header: "Ports",
        cell: (portsCell) => <PortsSummary ports={portsCell.getValue()} />,
        meta: { centerAligned: true },
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
    [enclaveShortUUID],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
