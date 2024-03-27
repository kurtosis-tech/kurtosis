import { Button, ButtonProps, Tooltip, useToast } from "@chakra-ui/react";
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
  const toast = useToast();

  const handleEnable = async () => {
    const createWebhookResponse = await createWebhook(packageId);
    if (createWebhookResponse.isOk) {
      toast({
        position: "bottom",
        title: "Enabled preview environments",
        colorScheme: "green",
      });
    } else {
      toast({
        position: "bottom",
        title: `Couldn't create preview envirionments and got error: ${createWebhookResponse.error}`,
        colorScheme: "red",
      });
    }
    setShowModal(false);
  };

  return (
    <>
      <Tooltip
        label={`This will create a webhook in your repository that enables per PR preview environments`}
        openDelay={1000}
      >
        <a
          href="https://github.com/apps/kurtosis-preview-environments/installations/select_target"
          target="_blank"
          rel="noopener noreferrer"
        >
          <Button
            colorScheme={"yellow"}
            leftIcon={<FiGithub />}
            // onClick={() => setShowModal(true)}
            size={"sm"}
            variant={"solid"}
            {...buttonProps}
          >
            Enable Preview Environments
          </Button>
        </a>
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
