import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useState } from "react";
import { FiEdit2 } from "react-icons/fi";
import { EnclaveFullInfo } from "../types";
import { ConfigureEnclaveModal } from "./modals/ConfigureEnclaveModal";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";

type EditEnclaveButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

export const EditEnclaveButton = ({ enclave, ...buttonProps }: EditEnclaveButtonProps) => {
  const [showPackageLoader, setShowPackageLoader] = useState(false);
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();

  const handlePackageLoaded = (kurtosisPackage: KurtosisPackage) => {
    setShowPackageLoader(false);
    setKurtosisPackage(kurtosisPackage);
  };

  if (!isDefined(enclave.starlarkRun)) {
    return (
      <Button isLoading={true} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"sm"} {...buttonProps}>
        Edit
      </Button>
    );
  }

  if (enclave.starlarkRun.isErr) {
    return (
      <Tooltip label={"Cannot find previous run config to edit"}>
        <Button isDisabled={true} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"sm"} {...buttonProps}>
          Edit
        </Button>
      </Tooltip>
    );
  }

  return (
    <>
      <Tooltip
        label={"Edit this enclave. From here you can edit the enclave configuration and update it."}
        openDelay={1000}
      >
        <Button
          onClick={() => setShowPackageLoader(true)}
          colorScheme={"blue"}
          leftIcon={<FiEdit2 />}
          size={"sm"}
          {...buttonProps}
        >
          Edit
        </Button>
      </Tooltip>
      {showPackageLoader && (
        <PackageLoadingModal packageId={enclave.starlarkRun.value.packageId} onPackageLoaded={handlePackageLoaded} />
      )}
      {isDefined(kurtosisPackage) && (
        <ConfigureEnclaveModal
          isOpen={true}
          onClose={() => setKurtosisPackage(undefined)}
          kurtosisPackage={kurtosisPackage}
          existingEnclave={enclave}
        />
      )}
    </>
  );
};
