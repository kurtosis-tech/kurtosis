import { Button, Tooltip } from "@chakra-ui/react";
import { useState } from "react";
import { FiEdit2 } from "react-icons/fi";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { isDefined } from "../../utils";
import { ConfigureEnclaveModal } from "./modals/ConfigureEnclaveModal";
import { PackageLoadingModal } from "./modals/PackageLoadingModal";

type EditEnclaveButtonProps = {
  enclave: EnclaveFullInfo;
};

export const EditEnclaveButton = ({ enclave }: EditEnclaveButtonProps) => {
  const [showPackageLoader, setShowPackageLoader] = useState(false);
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();

  const handlePackageLoaded = (kurtosisPackage: KurtosisPackage) => {
    setShowPackageLoader(false);
    setKurtosisPackage(kurtosisPackage);
  };

  if (!isDefined(enclave.starlarkRun)) {
    return (
      <Button isLoading={true} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"md"}>
        Edit
      </Button>
    );
  }

  if (enclave.starlarkRun.isErr) {
    return (
      <Tooltip label={"Cannot find previous run config to edit"}>
        <Button isDisabled={true} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"md"}>
          Edit
        </Button>
      </Tooltip>
    );
  }

  return (
    <>
      <Button onClick={() => setShowPackageLoader(true)} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"md"}>
        Edit
      </Button>
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
