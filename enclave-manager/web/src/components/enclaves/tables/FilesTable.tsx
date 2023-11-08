import { Button } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import {
  FilesArtifactNameAndUuid,
  ListFilesArtifactNamesAndUuidsResponse,
} from "enclave-manager-sdk/build/api_container_service_pb";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { RemoveFunctions } from "../../../utils/types";
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
  enclaveShortUUID: string;
  filesAndArtifacts: RemoveFunctions<ListFilesArtifactNamesAndUuidsResponse>;
};

export const FilesTable = ({ filesAndArtifacts, enclaveShortUUID }: FilesTableProps) => {
  const services = filesAndArtifacts.fileNamesAndUuids.map(fileToRow);

  const columns = useMemo<ColumnDef<FilesTableRow, any>[]>(
    () => [
      columnHelper.accessor("name", {
        header: "Name",
        cell: ({ row, getValue }) => (
          <Link to={`/enclave/${enclaveShortUUID}/file/${row.original.uuid}`}>
            <Button size={"sm"} variant={"ghost"}>
              {getValue()}
            </Button>
          </Link>
        ),
      }),
    ],
    [enclaveShortUUID],
  );

  return <DataTable columns={columns} data={services} defaultSorting={[{ id: "name", desc: true }]} />;
};
