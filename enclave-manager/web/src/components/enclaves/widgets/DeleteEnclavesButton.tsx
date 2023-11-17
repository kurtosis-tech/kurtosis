import { Button, Tooltip } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { FiTrash2 } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { useEmuiAppContext } from "../../../emui/EmuiAppContext";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { KurtosisAlertModal } from "../../KurtosisAlertModal";

type DeleteEnclavesButtonProps = {
  enclaves: EnclaveFullInfo[];
};

export const DeleteEnclavesButton = ({ enclaves }: DeleteEnclavesButtonProps) => {
  const { destroyEnclave } = useEmuiAppContext();
  const navigator = useNavigate();

  const [showModal, setShowModal] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const enclaveUUIDsKey = enclaves.map(({ enclaveUuid }) => enclaveUuid).join(",");

  useEffect(
    () => {
      setIsLoading(false);
      setShowModal(false);
    },
    // These deps are defined this way to detect whether or not the enclaves in props are actually different
    [enclaveUUIDsKey],
  );

  const handleDelete = async () => {
    setIsLoading(true);
    for (const enclaveUUID of enclaves.map(({ enclaveUuid }) => enclaveUuid)) {
      await destroyEnclave(enclaveUUID);
    }
    navigator("/enclaves");
    setIsLoading(false);
    setShowModal(false);
  };

  return (
    <>
      <Tooltip label={`This will delete ${enclaves.length} enclaves.`} openDelay={1000}>
        <Button colorScheme={"red"} leftIcon={<FiTrash2 />} onClick={() => setShowModal(true)}>
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
