import { Button, Checkbox } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { FilesArtifactNameAndUuid, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";
import { DateTime } from "luxon";
import { useMemo } from "react";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { DataTable } from "../DataTable";
import { RelativeDateTime } from "../RelativeDateTime";
import { EnclaveArtifactsSummary } from "./EnclaveArtifactsSummary";
import { EnclaveLinkButton } from "./EnclaveLinkButton";
import { EnclaveServicesSummary } from "./EnclaveServicesSummary";
import { EnclaveSourceButton } from "./EnclaveSourceButton";
import { EnclaveStatus } from "./EnclaveStatus";

type EnclaveTableRow = {
  uuid: string;
  name: string;
  status: EnclaveContainersStatus;
  created: DateTime | null;
  source: string;
  services: ServiceInfo[];
  artifacts: FilesArtifactNameAndUuid[];
};

const enclaveToRow = (enclave: EnclaveFullInfo): EnclaveTableRow => {
  return {
    uuid: enclave.enclaveUuid,
    name: enclave.name,
    status: enclave.containersStatus,
    created: enclave.creationTime ? DateTime.fromJSDate(enclave.creationTime.toDate()) : null,
    source: enclave.starlarkRun.packageId,
    services: Object.values(enclave.services.serviceInfo),
    artifacts: enclave.filesAndArtifacts.fileNamesAndUuids,
  };
};

const columnHelper = createColumnHelper<EnclaveTableRow>();

type EnclavesTableProps = {
  enclavesData: EnclaveFullInfo[];
  selection: EnclaveFullInfo[];
  onSelectionChange: (newSelection: EnclaveFullInfo[]) => void;
};

export const EnclavesTable = ({ enclavesData, selection, onSelectionChange }: EnclavesTableProps) => {
  const enclaves = enclavesData.map(enclaveToRow);

  const rowSelection = useMemo(() => {
    const selectedUUIDs = new Set<string>(selection.map(({ enclaveUuid }) => enclaveUuid));
    return enclavesData.reduce(
      (acc, cur, i) => {
        if (selectedUUIDs.has(cur.enclaveUuid)) {
          acc[i] = true;
        }
        // falsey values are not allowed - they break getIsSomeRowsSelected
        return acc;
      },
      {} as Record<string, boolean>,
    );
  }, [selection, enclavesData]);

  const columns = useMemo<ColumnDef<EnclaveTableRow, any>[]>(
    () => [
      columnHelper.accessor("uuid", {
        header: ({ table }) => (
          <Checkbox
            isIndeterminate={table.getIsSomeRowsSelected()}
            isChecked={table.getIsAllRowsSelected()}
            onChange={table.getToggleAllRowsSelectedHandler()}
          />
        ),
        cell: ({ row, getValue }) => (
          <Checkbox isChecked={row.getIsSelected()} onChange={row.getToggleSelectedHandler()} />
        ),
        enableSorting: false,
      }),
      columnHelper.accessor("name", {
        header: "Name",
        cell: (nameCell) => <EnclaveLinkButton name={nameCell.row.original.name} uuid={nameCell.row.original.uuid} />,
      }),
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
      columnHelper.accessor("artifacts", {
        cell: (artifactsCell) => <EnclaveArtifactsSummary artifacts={artifactsCell.getValue()} />,
      }),
    ],
    [],
  );

  return (
    <DataTable
      rowSelection={rowSelection}
      onRowSelectionChange={(updaterOrValue) => {
        const newRowSelection = typeof updaterOrValue === "function" ? updaterOrValue(rowSelection) : updaterOrValue;
        onSelectionChange(enclavesData.filter((enclave, i) => newRowSelection[i]));
      }}
      columns={columns}
      data={enclaves}
      defaultSorting={[{ id: "created", desc: true }]}
    />
  );
};
