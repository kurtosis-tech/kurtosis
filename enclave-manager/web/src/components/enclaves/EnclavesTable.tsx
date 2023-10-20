import { EnclaveInfo, GetEnclavesResponse } from "enclave-manager-sdk/build/engine_service_pb";
import { createColumnHelper } from "@tanstack/react-table";
import { DataTable } from "../DataTable";
import { isDefined } from "../../utils";
import { DateTime } from "luxon";
import { Tooltip } from "@chakra-ui/react";

const columnHelper = createColumnHelper<EnclaveInfo>();

const columns = [
  columnHelper.accessor("name", { header: "Name" }),
  columnHelper.accessor("apiContainerStatus", { header: "status" }),
  columnHelper.accessor("creationTime", {
    header: "Created",
    cell: (createdCell) => {
      const created = createdCell.getValue();
      if (!isDefined(created)) {
        return "Unknown";
      }
      const createdDateTime = DateTime.fromJSDate(created.toDate());
      return <Tooltip label={createdDateTime.toISO()}>{createdDateTime.toRelative()}</Tooltip>;
    },
    sortingFn: (r1, r2, cId) => {
      const d1 = (r1.getValue(cId) as any).toDate();
      const d2 = (r1.getValue(cId) as any).toDate();
      return d1 < d2 ? -1 : d1 === d2 ? 0 : 1;
    },
  }),
  columnHelper.accessor("en"),
];

//name, status, cerated, source, services, file artifact count
type EnclavesTableProps = {
  enclavesData: GetEnclavesResponse["enclaveInfo"];
};

export const EnclavesTable = ({ enclavesData }: EnclavesTableProps) => {
  const enclaves = Object.values(enclavesData);

  return <DataTable columns={columns} data={enclaves} defaultSorting={[{ id: "creationTime", desc: true }]} />;
};
