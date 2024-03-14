import {
  Button,
  Flex,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { CodeEditor, CodeEditorImperativeAttributes, isDefined } from "kurtosis-ui-components";
import { useMemo, useRef } from "react";

type EditFileModalProps = {
  isOpen: boolean;
  onClose: () => void;
  filePath: string[];
  file: string;
  onSave?: (newContents: string) => void;
};
export const EditFileModal = ({ isOpen, onClose, filePath, file, onSave }: EditFileModalProps) => {
  const codeEditorRef = useRef<CodeEditorImperativeAttributes>(null);

  const fileName = useMemo(() => "/" + filePath.join("/"), [filePath]);

  return (
    <Modal closeOnOverlayClick={true} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <ModalContent minW={"800px"} maxH={"90vh"}>
        <ModalHeader>
          Editing <b>{fileName}</b>
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody minH={"70vh"} p={"0"} bg={"gray.900"} flex={"1 1 auto"} overflowY={"auto"}>
          <CodeEditor ref={codeEditorRef} fileName={fileName} text={file} isEditable={isDefined(onSave)} />
        </ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button variant={"outline"} onClick={onClose}>
              Cancel
            </Button>
            {isDefined(onSave) && (
              <Button
                colorScheme={"kurtosisGreen"}
                variant={"outline"}
                onClick={() => onSave(codeEditorRef.current?.getText() || "")}
              >
                Save
              </Button>
            )}
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
