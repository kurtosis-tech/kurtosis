import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { useState } from "react";
import { FiGithub } from "react-icons/fi";
import { EnclaveFullInfo } from "../../types";
import { AddGithubActionModal } from "../modals/AddGithubActionModal";

type RunViaCIModalButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

export const AddGithubActionButton = ({ enclave, ...buttonProps }: RunViaCIModalButtonProps) => {
  const [showModal, setShowModal] = useState(false);

  if (!isDefined(enclave.starlarkRun)) {
    return (
      <Button isLoading={true} colorScheme={"yellow"} leftIcon={<FiGithub />} size={"sm"} {...buttonProps}>
        Add GitHub Action
      </Button>
    );
  }

  if (enclave.starlarkRun.isErr) {
    return (
      <Tooltip label={"This enclave didn't really load"}>
        <Button isDisabled={true} colorScheme={"yellow"} leftIcon={<FiGithub />} size={"sm"} {...buttonProps}>
          Add GitHub Action
        </Button>
      </Tooltip>
    );
  }

  return (
    <>
      <Tooltip label={`Steps to run this package from CI`} openDelay={1000}>
        <Button
          colorScheme={"yellow"}
          leftIcon={<FiGithub />}
          onClick={() => setShowModal(true)}
          size={"sm"}
          variant={"solid"}
          {...buttonProps}
        >
          Add GitHub Action
        </Button>
      </Tooltip>
      <AddGithubActionModal
        packageId={enclave.starlarkRun.value.packageId}
        isOpen={showModal}
        onClose={() => setShowModal(false)}
      />
    </>
  );
};
