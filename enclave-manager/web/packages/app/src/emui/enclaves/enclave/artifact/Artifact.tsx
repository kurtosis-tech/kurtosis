import { ButtonGroup, Flex, Spinner } from "@chakra-ui/react";
import { InspectFilesArtifactContentsResponse } from "enclave-manager-sdk/build/api_container_service_pb";
import {
  AppPageLayout,
  CodeEditor,
  CodeEditorImperativeAttributes,
  CopyButton,
  DownloadButton,
  FileTree,
  FileTreeNode,
  FormatButton,
  isDefined,
  KurtosisAlert,
  TitledCard,
} from "kurtosis-ui-components";
import { useEffect, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { Result } from "true-myth";
import { useKurtosisClient } from "../../../../client/enclaveManager/KurtosisClientContext";
import { useFullEnclave } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";

export const Artifact = () => {
  const { fileUUID, enclaveUUID } = useParams();

  if (!isDefined(fileUUID) || !isDefined(enclaveUUID)) {
    return (
      <AppPageLayout>
        <KurtosisAlert
          message={"Cannot load an artifact if the fileUUID or enclaveUUID are undefined - check the url"}
        />
      </AppPageLayout>
    );
  }

  return <ArtifactLoader enclaveUUID={enclaveUUID} fileUUID={fileUUID} />;
};

type ArtifactLoaderProps = {
  enclaveUUID: string;
  fileUUID: string;
};

const ArtifactLoader = ({ enclaveUUID, fileUUID }: ArtifactLoaderProps) => {
  const [filesResult, setFilesResult] = useState<Result<InspectFilesArtifactContentsResponse, string>>();

  const enclave = useFullEnclave(enclaveUUID);
  const kurtosisClient = useKurtosisClient();

  useEffect(() => {
    (async () => {
      if (enclave.isOk) {
        setFilesResult(undefined);
        const files = await kurtosisClient.inspectFilesArtifactContents(enclave.value, fileUUID);
        setFilesResult(files);
      }
    })();
  }, [kurtosisClient, enclave, fileUUID]);

  if (!isDefined(filesResult)) {
    return (
      <AppPageLayout>
        <Spinner />
      </AppPageLayout>
    );
  }

  if (filesResult.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={filesResult.error} />
      </AppPageLayout>
    );
  }

  if (enclave.isErr) {
    return (
      <AppPageLayout>
        <KurtosisAlert message={enclave.error} />
      </AppPageLayout>
    );
  }

  const artifactName =
    enclave.value.filesAndArtifacts?.mapOr(
      undefined,
      (files) => files.fileNamesAndUuids.find((file) => file.fileUuid === fileUUID)?.fileUuid,
    ) || "Unknown";

  return <ArtifactImpl files={filesResult.value} enclave={enclave.value} artifactName={artifactName} />;
};

type ArtifactImplProps = {
  enclave: EnclaveFullInfo;
  artifactName: string;
  files: InspectFilesArtifactContentsResponse;
};

const ArtifactImpl = ({ enclave, artifactName, files }: ArtifactImplProps) => {
  const codeEditorRef = useRef<CodeEditorImperativeAttributes>(null);
  const [selectedFilePath, setSelectedFilePath] = useState<string[]>();

  const filesAsFileTree = useMemo<FileTreeNode>(() => {
    return files.fileDescriptions
      .filter((fileDescription) => !fileDescription.path.endsWith("/"))
      .reduce(
        (acc, fileDescription): FileTreeNode => {
          const filePath = fileDescription.path.split("/");
          let destinationNode = acc;
          let i = 0;
          while (i < filePath.length - 1) {
            const filePart = filePath[i];
            let nextNode = destinationNode.childNodes?.find((node) => node.name === filePart);
            if (!isDefined(nextNode)) {
              nextNode = { name: filePart, childNodes: [] };
              destinationNode.childNodes?.push(nextNode);
            }
            destinationNode = nextNode;
            i++;
          }
          destinationNode.childNodes?.push({
            name: filePath[filePath.length - 1],
            size: fileDescription.size,
          });

          return acc;
        },
        { name: "root", childNodes: [] } as FileTreeNode,
      );
  }, [files]);

  const selectedFile = useMemo(() => {
    const path = selectedFilePath?.join("/");
    return files.fileDescriptions.find((file) => file.path === path);
  }, [files, selectedFilePath]);

  return (
    <AppPageLayout preventPageScroll>
      <Flex w={"100%"} h={"100%"} gap={"32px"} flex={"1 1 auto"}>
        <TitledCard title={"FILES"} w={"328px"} fillContainer>
          <Flex>
            <FileTree
              nodes={filesAsFileTree.childNodes || []}
              onFileSelected={setSelectedFilePath}
              selectedFilePath={selectedFilePath}
            />
          </Flex>
        </TitledCard>
        <TitledCard
          fillContainer
          title={isDefined(selectedFile) ? selectedFile.path : "Select a file to preview it"}
          controls={
            isDefined(selectedFile) ? (
              <ButtonGroup>
                <CopyButton
                  contentName={"File Path"}
                  isIconButton
                  aria-label={"Copy this file path"}
                  valueToCopy={selectedFile.path}
                />
                <DownloadButton
                  isIconButton
                  aria-label={"Download this file"}
                  valueToDownload={selectedFile.textPreview}
                  fileName={`${enclave.name}--${artifactName}-${selectedFile.path.replaceAll("/", "-")}`}
                />
              </ButtonGroup>
            ) : undefined
          }
          rightControls={
            isDefined(selectedFile) ? (
              <FormatButton variant="ghost" onClick={() => codeEditorRef.current?.formatCode()} />
            ) : undefined
          }
          flex={"1"}
          minH={"100%"}
        >
          {isDefined(selectedFile) && isDefined(selectedFilePath) && (
            <CodeEditor
              ref={codeEditorRef}
              // Use a key to force the editor to remount rather than mutate
              key={selectedFile.path}
              showLineNumbers
              text={selectedFile.textPreview || ""}
              fileName={selectedFilePath[(selectedFilePath?.length || 0) - 1]}
            />
          )}
        </TitledCard>
      </Flex>
    </AppPageLayout>
  );
};
