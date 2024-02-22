import { Button, ButtonGroup, Flex } from "@chakra-ui/react";
import { FileTree, FileTreeNode, isDefined } from "kurtosis-ui-components";
import { useMemo, useState } from "react";
import { Controller } from "react-hook-form";
import { FiPlus } from "react-icons/fi";
import { KurtosisFormInputProps } from "../../form/types";
import { EditFileModal } from "../modals/EditFileModal";
import { NewFileModal } from "../modals/NewFileModal";

type FileTreeArgumentInputProps<DataModel extends object> = KurtosisFormInputProps<DataModel>;

export const FileTreeArgumentInput = <DataModel extends object>({
  name,
  isRequired,
  validate,
  disabled,
}: FileTreeArgumentInputProps<DataModel>) => {
  return (
    <Controller
      name={name}
      disabled={disabled}
      defaultValue={"" as any}
      rules={{ required: isRequired, validate: validate }}
      render={({ field, fieldState }) => {
        return <FileTreeInput files={field.value} onUpdateFiles={field.onChange} />;
      }}
    />
  );
};

type FileTreeInputProps = {
  files: Record<string, string>;
  onUpdateFiles: (newFiles: Record<string, string>) => void;
};

const FileTreeInput = ({ files, onUpdateFiles }: FileTreeInputProps) => {
  const [selectedPath, setSelectedPath] = useState<string[]>();
  const [showNewFileInputDialog, setShowNewFileInputDialog] = useState(false);
  const [editingFilePath, setEditingFilePath] = useState<string[]>();

  const fileTree = useMemo((): FileTreeNode[] => {
    return (
      Object.entries(files).reduce(
        (acc, [fileName, fileContent]) => {
          let filePath = fileName.split("/");
          if (filePath[0] === "") {
            filePath = filePath.slice(1);
          }
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
            size: BigInt(fileContent.length),
          });

          return acc;
        },
        { name: "/", childNodes: [] } as FileTreeNode,
      ).childNodes || []
    );
  }, [files]);

  const handleNewFile = (newFileName: string) => {
    setShowNewFileInputDialog(false);
    onUpdateFiles({ ...files, [newFileName]: "" });
  };

  const handleSaveEditedFile = (text: string) => {
    if (!isDefined(editingFilePath)) {
      return;
    }
    onUpdateFiles({ ...files, ["/" + editingFilePath.join("/")]: text });
    setEditingFilePath(undefined);
  };

  const handleDeleteSelectedFile = () => {
    if (!isDefined(selectedPath)) {
      return;
    }
    const newFiles = { ...files };
    delete newFiles["/" + selectedPath.join("/")];
    onUpdateFiles(newFiles);
  };

  return (
    <Flex flexDirection={"column"} gap={"8px"}>
      <ButtonGroup size={"xs"} variant={"outline"}>
        <Button leftIcon={<FiPlus />} onClick={() => setShowNewFileInputDialog(true)} colorScheme={"kurtosisGreen"}>
          New File
        </Button>
        <Button onClick={handleDeleteSelectedFile} isDisabled={!isDefined(selectedPath)} colorScheme={"red"}>
          Delete
        </Button>
      </ButtonGroup>
      <FileTree
        nodes={fileTree}
        onFileSelected={setSelectedPath}
        selectedFilePath={selectedPath}
        onFileDblClicked={setEditingFilePath}
      />
      <NewFileModal
        isOpen={showNewFileInputDialog}
        onClose={() => setShowNewFileInputDialog(false)}
        onConfirm={handleNewFile}
      />
      <EditFileModal
        key={files["/" + editingFilePath?.join("/")]}
        isOpen={isDefined(editingFilePath)}
        onClose={() => setEditingFilePath(undefined)}
        filePath={editingFilePath || []}
        file={files["/" + editingFilePath?.join("/")]}
        onSave={handleSaveEditedFile}
      />
    </Flex>
  );
};
