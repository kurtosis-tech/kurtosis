import { Button, Checkbox, Text } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { ServiceInfo } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveContainersStatus } from "enclave-manager-sdk/build/engine_service_pb";
import { GetPackagesResponse, KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { DataTable, FormatDateTime, isDefined } from "kurtosis-ui-components";
import { DateTime } from "luxon";
import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Result } from "true-myth";
import { useKurtosisPackageIndexerClient } from "../../../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { EnclaveFullInfo } from "../../types";
import { EnclaveServicesSummary } from "../widgets/EnclaveServicesSummary";
import { EnclaveStatus } from "../widgets/EnclaveStatus";
import { PackageLinkButton } from "../widgets/PackageLinkButton";
import { PortsSummary } from "../widgets/PortsSummary";
import { getPortTableRows, PortsTableRow } from "./PortsTable";

type EnclaveTableRow = {
  uuid: string;
  name: string;
  status: EnclaveContainersStatus;
  created: DateTime | null;
  source: "loading" | KurtosisPackage | null;
  services: "loading" | ServiceInfo[] | null;
  ports: "loading" | PortsTableRow[] | null;
};

const enclaveToRow = (enclave: EnclaveFullInfo, catalog?: Result<GetPackagesResponse, string>): EnclaveTableRow => {
    const starlarkRun = enclave.starlarkRun;
  return {
    uuid: enclave.shortenedUuid,
    name: enclave.name,
    status: enclave.containersStatus,
    created: enclave.creationTime ? DateTime.fromJSDate(enclave.creationTime.toDate()) : null,
  source:
      !isDefined(starlarkRun) || !isDefined(catalog)
          ? "loading"
          : starlarkRun.isOk && catalog.isOk
              ? catalog.value.packages.find((kurtosisPackage) => kurtosisPackage.name === starlarkRun.value.packageId) || null
              : null,
    services: !isDefined(enclave.services)
      ? "loading"
      : enclave.services.isOk
      ? Object.values(enclave.services.value.serviceInfo)
      : null,
    ports: !isDefined(enclave.services)
      ? "loading"
      : enclave.services.isOk
      ? Object.values(enclave.services.value.serviceInfo).flatMap((service) =>
          getPortTableRows(
            enclave.enclaveUuid,
            service.serviceUuid,
            service.privatePorts,
            service.maybePublicPorts,
            service.maybePublicIpAddr,
            service.name,
          ),
        )
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
  const packageIndexerClient = useKurtosisPackageIndexerClient();
  const [catalog, setCatalog] = useState<Result<GetPackagesResponse, string>>();
  const enclaves = enclavesData.map((enclave) => enclaveToRow(enclave, catalog));

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
        cell: (sourceCell) => <PackageLinkButton source={sourceCell.getValue()} />,
      }),
      columnHelper.accessor("services", {
        cell: (servicesCell) => <EnclaveServicesSummary services={servicesCell.getValue()} />,
        meta: { centerAligned: true },
      }),
      columnHelper.accessor("ports", {
        header: "Ports",
        cell: (portsCell) => <PortsSummary disablePortLocking ports={portsCell.getValue()} />,
        meta: { centerAligned: true },
      }),
    ],
    [],
  );

  useEffect(() => {
    (async () => {
      setCatalog(await packageIndexerClient.getPackages());
    })();
  }, [packageIndexerClient]);

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
