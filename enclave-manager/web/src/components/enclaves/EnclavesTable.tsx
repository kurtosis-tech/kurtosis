import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";
import { createColumnHelper } from "@tanstack/react-table";
import { DataTable } from "../DataTable";
import { DateTime } from "luxon";
import { Button } from "@chakra-ui/react";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { EnclaveStatus } from "./EnclaveStatus";
import { EnclaveSourceButton } from "./EnclaveSourceButton";
import { RelativeDateTime } from "../RelativeDateTime";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveServicesSummary } from "./EnclaveServicesSummary";

type EnclaveTableRow = {
  name: string;
  status: EnclaveContainersStatus;
  created: DateTime | null;
  source: string;
  services: ServiceInfo[];
};

const enclaveToRow = (enclave: EnclaveFullInfo): EnclaveTableRow => {
  return {
    name: enclave.name,
    status: enclave.containersStatus,
    created: enclave.creationTime ? DateTime.fromJSDate(enclave.creationTime.toDate()) : null,
    source: enclave.starlarkRun.packageId,
    services: Object.values(enclave.services.serviceInfo),
  };
};

const columnHelper = createColumnHelper<EnclaveTableRow>();

const columns = [
  columnHelper.accessor("name", { header: "Name" }),
  columnHelper.accessor("status", {
    header: "Status",
    cell: (statusCell) => <EnclaveStatus status={statusCell.getValue()} />,
  }),
  columnHelper.accessor("created", {
    header: "Created",
    cell: (createdCell) => (
      <Button size={"xs"} variant={"kurtosisGhost"}>
        <RelativeDateTime dateTime={createdCell.getValue()} fontSize={""} />
      </Button>
    ),
  }),
  columnHelper.accessor("source", {
    header: "Source",
    cell: (sourceCell) => <EnclaveSourceButton source={sourceCell.getValue()} />,
  }),
  columnHelper.accessor("services", {
    cell: (servicesCell) => <EnclaveServicesSummary services={servicesCell.getValue()} />,
  }),
];

type EnclavesTableProps = {
  enclavesData: EnclaveFullInfo[];
};

export const EnclavesTable = ({ enclavesData }: EnclavesTableProps) => {
  const enclaves = Object.values(enclavesData).map(enclaveToRow);

  return <DataTable columns={columns} data={enclaves} defaultSorting={[{ id: "created", desc: true }]} />;
};
