import { Button, ButtonProps } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { FiPlay } from "react-icons/fi";
import { isDefined } from "../../utils";

type RunKurtosisPackageButtonProps = ButtonProps & {
  kurtosisPackage: KurtosisPackage;
};

export const RunKurtosisPackageButton = ({ kurtosisPackage, ...buttonProps }: RunKurtosisPackageButtonProps) => {
  return (
    <Button
      size={"xs"}
      variant={"solidOutline"}
      colorScheme={"kurtosisGreen"}
      leftIcon={<FiPlay />}
      {...buttonProps}
      onClick={(e) => {
        e.preventDefault();
        if (isDefined(buttonProps.onClick)) {
          buttonProps.onClick(e);
        }
      }}
    >
      Run
    </Button>
  );
};
