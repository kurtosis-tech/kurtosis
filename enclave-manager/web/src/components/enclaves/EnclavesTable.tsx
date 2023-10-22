import { EnclaveContainersStatus, EnclaveInfo, GetEnclavesResponse } from "enclave-manager-sdk/build/engine_service_pb";
import { createColumnHelper } from "@tanstack/react-table";
import { DataTable } from "../DataTable";
import { isDefined } from "../../utils";
import { DateTime } from "luxon";
import { Tooltip } from "@chakra-ui/react";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { EnclaveStatus } from "./EnclaveStatus";
import { EnclaveSource } from "./EnclaveSource";

type EnclaveTableRow = {
  name: string;
  status: EnclaveContainersStatus;
  created: DateTime | null;
  source: string;
};

const enclaveToRow = (enclave: EnclaveFullInfo): EnclaveTableRow => ({
  name: enclave.name,
  status: enclave.containersStatus,
  created: enclave.creationTime ? DateTime.fromJSDate(enclave.creationTime.toDate()) : null,
  source: enclave.starlarkRun.packageId,
});

const columnHelper = createColumnHelper<EnclaveTableRow>();

const columns = [
  columnHelper.accessor("name", { header: "Name" }),
  columnHelper.accessor("status", {
    header: "Status",
    cell: (statusCell) => <EnclaveStatus status={statusCell.getValue()} />,
  }),
  columnHelper.accessor("created", {
    header: "Created",
    cell: (createdCell) => {
      const created = createdCell.getValue();
      if (!isDefined(created)) {
        return "Unknown";
      }
      return <Tooltip label={created.toISO()}>{created.toRelative()}</Tooltip>;
    },
    sortingFn: (r1, r2, cId) => {
      const d1 = r1.getValue(cId) as any;
      const d2 = r1.getValue(cId) as any;
      return d1 < d2 ? -1 : d1 === d2 ? 0 : 1;
    },
  }),
  columnHelper.accessor("source", {
    header: "Source",
    cell: (sourceCell) => <EnclaveSource source={sourceCell.getValue()} />,
  }),
];

type EnclavesTableProps = {
  enclavesData: EnclaveFullInfo[];
};

export const EnclavesTable = ({ enclavesData }: EnclavesTableProps) => {
  const enclaves = Object.values(enclavesData).map(enclaveToRow);

  return <DataTable columns={columns} data={enclaves} defaultSorting={[{ id: "created", desc: true }]} />;
};
