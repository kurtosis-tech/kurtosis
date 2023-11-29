import { Button, ButtonProps } from "@chakra-ui/react";
import { useState } from "react";
import { FiDownload } from "react-icons/fi";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { EnclavesContextProvider } from "../../../emui/enclaves/EnclavesContext";
import { ConfigureEnclaveModal } from "../../enclaves/modals/ConfigureEnclaveModal";

type RunKurtosisPackageButtonProps = ButtonProps & {
  kurtosisPackage: KurtosisPackage;
};

export const RunKurtosisPackageButton = ({ kurtosisPackage, ...buttonProps }: RunKurtosisPackageButtonProps) => {
  const [configuringEnclave, setConfiguringEnclave] = useState(false);

  return (
    <>
      <Button
        size={"xs"}
        colorScheme={"kurtosisGreen"}
        leftIcon={<FiDownload />}
        onClick={() => setConfiguringEnclave(true)}
        {...buttonProps}
      >
        Run
      </Button>
      {configuringEnclave && (
        <EnclavesContextProvider skipInitialLoad>
          <ConfigureEnclaveModal
            isOpen={true}
            onClose={() => setConfiguringEnclave(false)}
            kurtosisPackage={kurtosisPackage}
          />
        </EnclavesContextProvider>
      )}
    </>
  );
};
