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
import { FileDisplay, isDefined } from "kurtosis-ui-components";

type ViewStarlarkModalProps = {
  isOpen: boolean;
  onClose: () => void;
  starlark?: string;
};
export const ViewStarlarkModal = ({ isOpen, onClose, starlark }: ViewStarlarkModalProps) => {
  return (
    <Modal closeOnOverlayClick={true} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <ModalContent minW={"800px"} maxH={"90vh"}>
        <ModalHeader>Previewing Starlark</ModalHeader>
        <ModalCloseButton />
        <ModalBody minH={"70vh"} flex={"1 1 auto"} overflowY={"auto"}>
          {isDefined(starlark) && <FileDisplay value={starlark} title={"main.star"} filename={`main.star`} />}
        </ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button variant={"outline"} onClick={onClose}>
              Close
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
