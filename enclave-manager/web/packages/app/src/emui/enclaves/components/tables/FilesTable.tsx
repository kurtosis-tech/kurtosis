import { Button } from "@chakra-ui/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import {
  FilesArtifactNameAndUuid,
  ListFilesArtifactNamesAndUuidsResponse,
} from "enclave-manager-sdk/build/api_container_service_pb";
import { DataTable, RemoveFunctions } from "kurtosis-ui-components";
import { useMemo } from "react";
import { Link } from "react-router-dom";
import { EnclaveFullInfo } from "../../types";
import { DownloadFileArtifactButton } from "../widgets/DownloadFileArtifactButton";

const columnHelper = createColumnHelper<FilesArtifactNameAndUuid>();

type FilesTableProps = {
  enclave: EnclaveFullInfo;
  filesAndArtifacts: RemoveFunctions<ListFilesArtifactNamesAndUuidsResponse>;
};

export const FilesTable = ({ filesAndArtifacts, enclave }: FilesTableProps) => {
  const columns = useMemo<ColumnDef<FilesArtifactNameAndUuid, any>[]>(
    () => [
      columnHelper.accessor("fileName", {
        header: "Name",
        cell: ({ row, getValue }) => (
          <Link to={`/enclave/${enclave.shortenedUuid}/file/${row.original.fileUuid}`}>
            <Button size={"sm"} variant={"ghost"}>
              {getValue()}
            </Button>
          </Link>
        ),
      }),
      columnHelper.display({
        id: "download",
        cell: ({ row }) => <DownloadFileArtifactButton file={row.original} enclave={enclave} />,
      }),
    ],
    [enclave],
  );

  return (
    <DataTable
      columns={columns}
      data={filesAndArtifacts.fileNamesAndUuids}
      defaultSorting={[{ id: "fileName", desc: true }]}
    />
  );
};
