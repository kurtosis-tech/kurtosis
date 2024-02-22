import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { useState } from "react";
import { FiLink2 } from "react-icons/fi";
import { EnclaveFullInfo } from "../../types";
import { ConnectEnclaveModal } from "../modals/ConnectEnclaveModal";

type ConnectEnclaveButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

export const ConnectEnclaveButton = ({ enclave, ...buttonProps }: ConnectEnclaveButtonProps) => {
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
      <ConnectEnclaveModal enclave={enclave} isOpen={showModal} onClose={() => setShowModal(false)} />
    </>
  );
};
