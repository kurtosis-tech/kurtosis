import { KurtosisAlertModal } from "./KurtosisAlertModal";

type FeatureNotImplementedModalProps = {
  featureName: string;
  issueUrl: string;
  message?: string;
  isOpen: boolean;
  onClose: () => void;
};

export const FeatureNotImplementedModal = ({
  featureName,
  issueUrl,
  message,
  isOpen,
  onClose,
}: FeatureNotImplementedModalProps) => {
  return (
    <KurtosisAlertModal
      title={`${featureName} unavailable`}
      isOpen={isOpen}
      onClose={onClose}
      confirmText={"Go to Issue"}
      onConfirm={() => {
        onClose();
        window.open(issueUrl, "_blank");
      }}
      confirmButtonProps={{ colorScheme: "kurtosisGreen" }}
      content={
        message ||
        `${featureName} is not currently available. Please comment/upvote the issue if you would like to use it.`
      }
    />
  );
};
