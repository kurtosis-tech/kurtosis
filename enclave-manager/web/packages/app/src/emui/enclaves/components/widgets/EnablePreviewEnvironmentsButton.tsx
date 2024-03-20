import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { KurtosisAlertModal } from "kurtosis-ui-components";
import { useState } from "react";
import { FiGithub } from "react-icons/fi";
import { useEnclavesContext } from "../../EnclavesContext";

type EnablePreviewEnvironmentsButtonProps = ButtonProps & {
  packageId: string;
};

export const EnablePreviewEnvironmentsButton = ({
  packageId,
  ...buttonProps
}: EnablePreviewEnvironmentsButtonProps) => {
  const { createWebhook } = useEnclavesContext();
  const [showModal, setShowModal] = useState(false);

  const handleEnable = async () => {
    await createWebhook(packageId);
    setShowModal(false);
  };

  return (
    <>
      <Tooltip
        label={`This will create a webhook in your repository that enables per PR preview environments`}
        openDelay={1000}
      >
        <Button
          colorScheme={"green"}
          leftIcon={<FiGithub />}
          onClick={() => setShowModal(true)}
          size={"sm"}
          variant={"solid"}
          {...buttonProps}
        >
          Enable Preview Environments
        </Button>
      </Tooltip>
      <KurtosisAlertModal
        isOpen={showModal}
        title={"Enable preview environments"}
        content={"This will enable preview environments on your repository per PR"}
        confirmText={"Enable"}
        confirmButtonProps={{ leftIcon: <FiGithub />, colorScheme: "green" }}
        onClose={() => setShowModal(false)}
        onConfirm={handleEnable}
      />
    </>
  );
};
