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

export type LogNavigationWarningModalProps = {
  isOpen: boolean;
  onCancel: () => void;
  onConfirm: () => void;
};

export const LogNavigationWarningModal = ({ isOpen, onConfirm, onCancel }: LogNavigationWarningModalProps) => {
  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onCancel} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Warning</ModalHeader>
        <ModalCloseButton />
        <ModalBody>You will not be able to access these logs after leaving this view.</ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button variant={"outline"} onClick={onCancel}>
              Cancel
            </Button>
            <Button colorScheme={"kurtosisGreen"} variant={"outline"} onClick={onConfirm}>
              Continue
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
