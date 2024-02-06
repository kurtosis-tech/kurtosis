import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { KurtosisAlertModal } from "kurtosis-ui-components";
import { useState } from "react";
import { FiTrash2 } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { useEnclavesContext } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";

type DeleteEnclavesButtonProps = ButtonProps & {
  enclaves: EnclaveFullInfo[];
};

export const DeleteEnclavesButton = ({ enclaves, ...buttonProps }: DeleteEnclavesButtonProps) => {
  const { destroyEnclaves } = useEnclavesContext();
  const navigator = useNavigate();

  const [showModal, setShowModal] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const handleDelete = async () => {
    setIsLoading(true);
    await destroyEnclaves(enclaves.map(({ enclaveUuid }) => enclaveUuid));
    navigator("/enclaves");
    setIsLoading(false);
    setShowModal(false);
  };

  return (
    <>
      <Tooltip label={`This will delete ${enclaves.length} enclaves.`} openDelay={1000}>
        <Button
          colorScheme={"red"}
          leftIcon={<FiTrash2 />}
          onClick={() => setShowModal(true)}
          size={"sm"}
          variant={"solid"}
          {...buttonProps}
        >
          Delete
        </Button>
      </Tooltip>
      <KurtosisAlertModal
        isOpen={showModal}
        isLoading={isLoading}
        title={"Delete Enclaves"}
        content={"Are you sure? You cannot undo this action afterwards."}
        confirmText={"Delete"}
        confirmButtonProps={{ leftIcon: <FiTrash2 />, colorScheme: "red" }}
        onClose={() => setShowModal(false)}
        onConfirm={handleDelete}
      />
    </>
  );
};
