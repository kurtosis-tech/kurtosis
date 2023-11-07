import { KurtosisAlertModal } from "./KurtosisAlertModal";

type FeatureNotImplementedModalProps = {
  featureName: string;
  message?: string;
  isOpen: boolean;
  onClose: () => void;
};

export const FeatureNotImplementedModal = ({
  featureName,
  message,
  isOpen,
  onClose,
}: FeatureNotImplementedModalProps) => {
  return (
    <KurtosisAlertModal
      title={`${featureName} unavailable`}
      isOpen={isOpen}
      onClose={onClose}
      confirmText={"Submit Request"}
      onConfirm={() => {
        onClose();
        window.open("https://github.com/kurtosis-tech/kurtosis/issues", "_blank");
      }}
      confirmButtonProps={{ colorScheme: "kurtosisGreen" }}
      content={
        message || `${featureName} is not currently available. Please open a feature request if you'd like to use this.`
      }
    />
  );
};
