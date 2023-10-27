import { Button } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { DataTable } from "../../DataTable";

type FilesTableRow = {
  name: string;
  //size: string; TODO: Add size to FilesArtifactNameAndUuid
  //description: string; TODO: Add description to FilesArtifactNameAndUuid
  uuid: string;
};

const fileToRow = (file: FilesArtifactNameAndUuid): FilesTableRow => {
  return {
    name: file.fileName,
    uuid: file.fileUuid,
  };
};

const columnHelper = createColumnHelper<FilesTableRow>();

type FilesTableProps = {
  enclave: EnclaveFullInfo;
};

export const FilesTable = ({ enclave }: FilesTableProps) => {
  const services = enclave.filesAndArtifacts.fileNamesAndUuids.map(fileToRow);

  const columns = useMemo<ColumnDef<FilesTableRow, any>[]>(
    () => [
      columnHelper.accessor("name", {
        header: "Name",
        cell: ({ row, getValue }) => (
          <Link to={`/enclave/${enclave.enclaveUuid}/file/${row.original.uuid}`}>
            <Button size={"sm"} variant={"ghost"}>
              {getValue()}
            </Button>
          </Link>
        ),
      }),
    ],
    [],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
