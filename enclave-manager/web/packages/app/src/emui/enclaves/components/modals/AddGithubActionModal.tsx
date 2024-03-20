import {
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { FileDisplay, KurtosisAlertModal } from "kurtosis-ui-components";
import { useState } from "react";
import { FiGithub } from "react-icons/fi";
import { apiKey, instanceUUID, isPrevEnv } from "../../../../cookies";
import { useEnclavesContext } from "../../EnclavesContext";

export type AddGithubActionModalProps = {
  packageId: string;
  isOpen: boolean;
  onClose: () => void;
};

export const AddGithubActionModal = ({ isOpen, onClose, packageId }: AddGithubActionModalProps) => {
  const { createWebhook } = useEnclavesContext();
  const [showModal, setShowModal] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [loadError, setLoadError] = useState<string>();

  // TODO handle failure
  // TODO create own modal+button and put that behind condition
  if (isPrevEnv) {
    const handleEnable = async () => {
      setIsLoading(true);
      const webhookResponse = await createWebhook(packageId);
      if (webhookResponse.isErr) {
        setLoadError("An error occurred while creating webhook");
      }
      setIsLoading(false);
      setShowModal(false);
    };

    return (
      <KurtosisAlertModal
        isOpen={showModal}
        isLoading={isLoading}
        title={"Enable preview environments"}
        content={"This will enable preview environments on your repository per PR"}
        confirmText={"Enable"}
        confirmButtonProps={{ leftIcon: <FiGithub />, colorScheme: "green" }}
        onClose={() => setShowModal(false)}
        onConfirm={handleEnable}
      />
    );
  }

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
