import { Button, ButtonProps } from "@chakra-ui/react";
import { useState } from "react";
import { FiPlay } from "react-icons/fi";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { EnclavesContextProvider } from "../../../emui/enclaves/EnclavesContext";
import { ConfigureEnclaveModal } from "../../enclaves/modals/ConfigureEnclaveModal";

type RunKurtosisPackageButtonProps = ButtonProps & {
  kurtosisPackage: KurtosisPackage;
};

export const RunKurtosisPackageButton = ({ kurtosisPackage, ...buttonProps }: RunKurtosisPackageButtonProps) => {
  const [isConfiguringEnclave, setIsConfiguringEnclave] = useState(false);

  return (
    <>
      <Button
        size={"xs"}
        variant={"solidOutline"}
        colorScheme={"kurtosisGreen"}
        leftIcon={<FiPlay />}
        onClick={(e) => {
          e.preventDefault();
          setIsConfiguringEnclave(true);
        }}
        isActive={isConfiguringEnclave}
        isLoading={isConfiguringEnclave}
        loadingText={"Configuring"}
        {...buttonProps}
      >
        Run
      </Button>
      {isConfiguringEnclave && (
        <EnclavesContextProvider skipInitialLoad>
          <ConfigureEnclaveModal
            isOpen={true}
            onClose={() => setIsConfiguringEnclave(false)}
            kurtosisPackage={kurtosisPackage}
          />
        </EnclavesContextProvider>
      )}
    </>
  );
};
