import {
  Button,
  ButtonProps,
  Flex,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
} from "@chakra-ui/react";

type KurtosisAlertModalProps = {
  title: string;
  content: string;
  isOpen: boolean;
  isLoading?: boolean;
  onClose: () => void;
  onConfirm: () => void;
  confirmText: string;
  confirmButtonProps?: ButtonProps;
};

export const KurtosisAlertModal = ({
  title,
  content,
  isOpen,
  isLoading,
  onClose,
  onConfirm,
  confirmText,
  confirmButtonProps,
}: KurtosisAlertModalProps) => {
  return (
    <Modal isOpen={isOpen} onClose={() => !isLoading && onClose()} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>{title}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Text>{content}</Text>
        </ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button color={"gray.100"} onClick={onClose} isDisabled={isLoading}>
              Dismiss
            </Button>
            <Button onClick={onConfirm} {...confirmButtonProps} isLoading={isLoading}>
              {confirmText}
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
