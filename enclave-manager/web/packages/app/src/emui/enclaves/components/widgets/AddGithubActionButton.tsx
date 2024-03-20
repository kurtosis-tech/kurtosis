import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { useState } from "react";
import { FiGithub } from "react-icons/fi";
import { isPrevEnv } from "../../../../cookies";
import { EnclaveFullInfo } from "../../types";
import { AddGithubActionModal } from "../modals/AddGithubActionModal";

type AddGithubActionButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

export const AddGithubActionButton = ({ enclave, ...buttonProps }: AddGithubActionButtonProps) => {
  const [showModal, setShowModal] = useState(false);

  let tooltip = "Add GitHub Action";
  if (isPrevEnv) {
    tooltip = "Enable Preview Envirionments";
  }

  if (!isDefined(enclave.starlarkRun)) {
    return (
      <Button isLoading={true} colorScheme={"yellow"} leftIcon={<FiGithub />} size={"sm"} {...buttonProps}>
        {tooltip}
      </Button>
    );
  }

  if (enclave.starlarkRun.isErr) {
    return (
      <Tooltip label={"An error occurred while starting the enclave"}>
        <Button isDisabled={true} colorScheme={"yellow"} leftIcon={<FiGithub />} size={"sm"} {...buttonProps}>
          {tooltip}
        </Button>
      </Tooltip>
    );
  }

  return (
    <>
      <Tooltip label={tooltip} openDelay={1000}>
        <Button
          colorScheme={"yellow"}
          leftIcon={<FiGithub />}
          onClick={() => setShowModal(true)}
          size={"sm"}
          variant={"solid"}
          {...buttonProps}
        >
          {tooltip}
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
