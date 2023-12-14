import { FilesArtifactNameAndUuid } from "enclave-manager-sdk/build/api_container_service_pb";
import { DownloadButton } from "kurtosis-ui-components";
import { useState } from "react";
import streamsaver from "streamsaver";
import { useKurtosisClient } from "../../../../client/enclaveManager/KurtosisClientContext";
import { EnclaveFullInfo } from "../../types";

type DownloadFileButtonProps = {
  file: FilesArtifactNameAndUuid;
  enclave: EnclaveFullInfo;
};

export const DownloadFileArtifactButton = ({ file, enclave }: DownloadFileButtonProps) => {
  const kurtosisClient = useKurtosisClient();
  const [isLoading, setIsLoading] = useState(false);

  const handleDownloadClick = async () => {
    setIsLoading(true);
    const fileParts = await kurtosisClient.downloadFilesArtifact(enclave, file);
    const writableStream = streamsaver.createWriteStream(`${enclave.name}--${file.fileName}.tgz`);
    const writer = writableStream.getWriter();

    for await (const part of fileParts) {
      await writer.write(part.data);
    }
    await writer.close();
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
