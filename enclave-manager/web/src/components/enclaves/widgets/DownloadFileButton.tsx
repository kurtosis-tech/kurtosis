import { useToast } from "@chakra-ui/react";
import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { useState } from "react";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { saveTextAsFile } from "../../../utils/download";
import { DownloadButton } from "../../DownloadButton";

type DownloadFileButtonProps = {
  file: FilesArtifactNameAndUuid;
  enclave: EnclaveFullInfo;
};

export const DownloadFileButton = ({ file, enclave }: DownloadFileButtonProps) => {
  const kurtosisClient = useKurtosisClient();
  const toast = useToast();
  const [isLoading, setIsLoading] = useState(false);

  const handleDownloadClick = async () => {
    setIsLoading(true);
    // todo: get tgz download instead
    const maybeFile = await kurtosisClient.inspectFilesArtifactContents(enclave, file);
    if (maybeFile.isErr) {
      toast({
        title: `Could not inspect ${file.fileName}: ${maybeFile.error}`,
        colorScheme: "red",
      });
      setIsLoading(false);
      return;
    }

    saveTextAsFile("some file", `${enclave.name}-${file.fileName}.tgz`);
    setIsLoading(false);
  };

  return (
    <DownloadButton
      fileName={file.fileName}
      isIconButton
      aria-label={`Download ${file.fileName}`}
      isLoading={isLoading}
      onClick={handleDownloadClick}
    />
  );
};
