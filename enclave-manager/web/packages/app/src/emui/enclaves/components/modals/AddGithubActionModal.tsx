import {
  ButtonProps,
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

export type AddGithubActionModalProps = ButtonProps & {
  packageId: string;
  isOpen: boolean;
  onClose: () => void;
};

export const AddGithubActionModal = ({ isOpen, onClose, packageId, ...buttonProps }: AddGithubActionModalProps) => {
  const commands = `
name: CI
on:
    pull_request:

jobs:
  run_kurtosis:
    runs-on: ubuntu-latest
      steps:
        - name: Checkout Repository
          uses: actions/checkout@v4
        - name: Run Kurtosis
          uses: kurtosis-tech/kurtosis-github-action@v1
          with:
          path: ${packageId}
          cloud_instance_id: ${instanceUUID}
          cloud_api_key: ${apiKey} # We recommend placing this in repository secrets;`;
  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Run this enclave from GitHub</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <FileDisplay value={commands} title={"GitHub Action YAML"} filename={"per-pr.yml"} />
        </ModalBody>
        <ModalFooter>The GitHub Action allows you to run this package directly in CI</ModalFooter>
      </ModalContent>
    </Modal>
  );
};
