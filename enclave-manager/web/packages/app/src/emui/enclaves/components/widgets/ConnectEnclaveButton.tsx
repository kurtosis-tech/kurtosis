import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { useState } from "react";
import { FiLink2 } from "react-icons/fi";
import { EnclaveFullInfo } from "../../types";
import { ConnectEnclaveModal } from "../modals/ConnectEnclaveModal";

type ConnectEnclaveButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
  instanceUUID: string;
};

export const ConnectEnclaveButton = ({ enclave, instanceUUID, ...buttonProps }: ConnectEnclaveButtonProps) => {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <Tooltip label={`Steps to connect to this enclave from the CLI.`} openDelay={1000}>
        <Button
          colorScheme={"green"}
          leftIcon={<FiLink2 />}
          onClick={() => setShowModal(true)}
          size={"sm"}
          variant={"solid"}
          {...buttonProps}
        >
          Connect
        </Button>
      </Tooltip>
      <ConnectEnclaveModal
        enclave={enclave}
        instanceUUID={instanceUUID}
        isOpen={showModal}
        onClose={() => setShowModal(false)}
      />
    </>
  );
};
