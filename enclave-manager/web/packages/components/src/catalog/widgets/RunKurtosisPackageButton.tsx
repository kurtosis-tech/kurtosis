import { Button, ButtonProps } from "@chakra-ui/react";
import { FiPlay } from "react-icons/fi";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../../utils";

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
