import {
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { FileDisplay } from "kurtosis-ui-components";
import { apiKey, instanceUUID } from "../../../../cookies";
import { EnclaveFullInfo } from "../../types";

export type ConnectEnclaveModalProps = {
  enclave: EnclaveFullInfo;
  isOpen: boolean;
  onClose: () => void;
};

export const ConnectEnclaveModal = ({ isOpen, onClose, enclave }: ConnectEnclaveModalProps) => {
  const commands = `
  export KURTOSIS_CLOUD_API_KEY="${apiKey}"
  kurtosis cloud load ${instanceUUID}
  kurtosis enclave connect ${enclave.name}
  kurtosis enclave inspect ${enclave.name}`;
  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Connect to this enclave from the CLI</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <FileDisplay value={commands} title={"CLI Commands"} filename={`${enclave.name}--connect.sh`} />
        </ModalBody>
        <ModalFooter>
          The enclave inspect command shows you the ephemeral ports to use to connect to your user services.
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
