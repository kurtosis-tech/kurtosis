import { Button, ButtonProps, Tooltip } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useState } from "react";
import { FiEdit2 } from "react-icons/fi";
import { useSettings } from "../../settings";
import { EnclaveFullInfo } from "../types";
import { CreateOrConfigureEnclaveDrawer } from "./configuration/drawer/CreateOrConfigureEnclaveDrawer";
import { EnclaveBuilderModal } from "./enclaveBuilder/EnclaveBuilderModal";
import { starlarkScriptContainsEMUIBuildState } from "./enclaveBuilder/utils";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";

type EditEnclaveButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

export const EditEnclaveButton = ({ enclave, ...buttonProps }: EditEnclaveButtonProps) => {
  const { settings } = useSettings();

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

  if (
    settings.ENABLE_EXPERIMENTAL_BUILD_ENCLAVE &&
    starlarkScriptContainsEMUIBuildState(enclave.starlarkRun.value.serializedScript)
  ) {
    return <EditFromScriptButton enclave={enclave} {...buttonProps} />;
  }

  return <EditFromPackageButton enclave={enclave} packageId={enclave.starlarkRun.value.packageId} {...buttonProps} />;
};

type EditFromPackageButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
  packageId: string;
};

const EditFromPackageButton = ({ enclave, packageId, ...buttonProps }: EditFromPackageButtonProps) => {
  const [showPackageLoader, setShowPackageLoader] = useState(false);
  const [showEnclaveConfiguration, setShowEnclaveConfiguration] = useState(false);
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();

  const handlePackageLoaded = (kurtosisPackage: KurtosisPackage) => {
    setShowPackageLoader(false);
    setKurtosisPackage(kurtosisPackage);
    setShowEnclaveConfiguration(true);
  };

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
      {showPackageLoader && <PackageLoadingModal packageId={packageId} onPackageLoaded={handlePackageLoaded} />}
      <CreateOrConfigureEnclaveDrawer
        isOpen={showEnclaveConfiguration}
        onClose={() => {
          setKurtosisPackage(undefined);
          setShowEnclaveConfiguration(false);
        }}
        kurtosisPackage={kurtosisPackage}
        existingEnclave={enclave}
      />
    </>
  );
};

type EditFromScriptButtonProps = ButtonProps & {
  enclave: EnclaveFullInfo;
};

const EditFromScriptButton = ({ enclave, ...buttonProps }: EditFromScriptButtonProps) => {
  const [showBuilderModal, setShowBuilderModal] = useState(false);

  return (
    <>
      <Tooltip
        label={"Edit this enclave. From here you can edit the enclave configuration and update it."}
        openDelay={1000}
      >
        <Button
          onClick={() => setShowBuilderModal(true)}
          colorScheme={"blue"}
          leftIcon={<FiEdit2 />}
          size={"sm"}
          {...buttonProps}
        >
          Edit
        </Button>
      </Tooltip>
      <EnclaveBuilderModal
        isOpen={showBuilderModal}
        onClose={() => setShowBuilderModal(false)}
        existingEnclave={enclave}
      />
    </>
  );
};
