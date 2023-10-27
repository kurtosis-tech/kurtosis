import { Button } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { FiTrash2 } from "react-icons/fi";
import { useFetcher } from "react-router-dom";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { KurtosisAlertModal } from "../../KurtosisAlertModal";

type DeleteEnclavesButtonProps = {
  enclaves: EnclaveFullInfo[];
};

export const DeleteEnclavesButton = ({ enclaves }: DeleteEnclavesButtonProps) => {
  const fetcher = useFetcher();

  const [showModal, setShowModal] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    setIsLoading(false);
    setShowModal(false);
  }, [enclaves]);

  const handleDelete = async () => {
    setIsLoading(true);
    fetcher.submit(
      { intent: "delete", enclaveUUIDs: enclaves.map(({ enclaveUuid }) => enclaveUuid) },
      { method: "post", action: "/enclaves", encType: "application/json" },
    );
  };

  return (
    <>
      <Button colorScheme={"red"} leftIcon={<FiTrash2 />} onClick={() => setShowModal(true)}>
        Delete
      </Button>
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
