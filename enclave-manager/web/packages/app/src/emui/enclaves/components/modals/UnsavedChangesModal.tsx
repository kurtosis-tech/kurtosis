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

export type UnsavedChangesModalProps = {
  isOpen: boolean;
  onCancel: () => void;
  onConfirm: () => void;
};

export const UnsavedChangesModal = ({ isOpen, onConfirm, onCancel }: UnsavedChangesModalProps) => {
  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onCancel} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>You have unsaved changes</ModalHeader>
        <ModalCloseButton />
        <ModalBody>Do you want to discard them?</ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button variant={"outline"} onClick={onCancel}>
              Cancel
            </Button>
            <Button colorScheme={"red"} variant={"outline"} onClick={onConfirm}>
              Yes, discard
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
