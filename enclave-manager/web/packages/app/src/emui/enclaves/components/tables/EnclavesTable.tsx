import { Button, Checkbox, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { FilesArtifactNameAndUuid, ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";
import { DataTable, FormatDateTime, isDefined, PackageSourceButton } from "kurtosis-ui-components";
import { DateTime } from "luxon";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { EnclaveFullInfo } from "../../types";
import { EnclaveArtifactsSummary } from "../widgets/EnclaveArtifactsSummary";
import { EnclaveServicesSummary } from "../widgets/EnclaveServicesSummary";
import { EnclaveStatus } from "../widgets/EnclaveStatus";

type EnclaveTableRow = {
  uuid: string;
  name: string;
  status: EnclaveContainersStatus;
  created: DateTime | null;
  source: "loading" | string | null;
  services: "loading" | ServiceInfo[] | null;
  artifacts: "loading" | FilesArtifactNameAndUuid[] | null;
};

const enclaveToRow = (enclave: EnclaveFullInfo): EnclaveTableRow => {
  return {
    uuid: enclave.shortenedUuid,
    name: enclave.name,
    status: enclave.containersStatus,
    created: enclave.creationTime ? DateTime.fromJSDate(enclave.creationTime.toDate()) : null,
    source: !isDefined(enclave.starlarkRun)
      ? "loading"
      : enclave.starlarkRun.isOk
      ? enclave.starlarkRun.value.packageId
      : null,
    services: !isDefined(enclave.services)
      ? "loading"
      : enclave.services.isOk
      ? Object.values(enclave.services.value.serviceInfo)
      : null,
    artifacts: !isDefined(enclave.filesAndArtifacts)
      ? "loading"
      : enclave.filesAndArtifacts.isOk
      ? enclave.filesAndArtifacts.value.fileNamesAndUuids
      : null,
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
        cell: (nameCell) => (
          <Link to={`/enclave/${nameCell.row.original.uuid}/overview`}>
            <Button size={"sm"} variant={"ghost"}>
              <Text as={"span"} maxW={"200px"} textOverflow={"ellipsis"} overflow={"hidden"}>
                {nameCell.row.original.name}
              </Text>
            </Button>
          </Link>
        ),
      }),
      columnHelper.accessor("status", {
        header: "Status",
        cell: (statusCell) => <EnclaveStatus status={statusCell.getValue()} variant={"square"} />,
      }),
      columnHelper.accessor("created", {
        header: "Created",
        cell: (createdCell) => (
          <FormatDateTime
            fontSize={"xs"}
            fontWeight={"semibold"}
            dateTime={createdCell.getValue()}
            format={"relative"}
          />
        ),
      }),
      columnHelper.accessor("source", {
        header: "Source",
        cell: (sourceCell) => <PackageSourceButton source={sourceCell.getValue()} />,
      }),
      columnHelper.accessor("services", {
        cell: (servicesCell) => <EnclaveServicesSummary services={servicesCell.getValue()} />,
        meta: { centerAligned: true },
      }),
      columnHelper.accessor("artifacts", {
        header: "File artifacts",
        cell: (artifactsCell) => <EnclaveArtifactsSummary artifacts={artifactsCell.getValue()} />,
        meta: { centerAligned: true },
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
