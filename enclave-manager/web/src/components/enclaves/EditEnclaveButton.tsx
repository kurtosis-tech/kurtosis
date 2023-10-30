import { Button, Tooltip } from "@chakra-ui/react";
import { EnclaveMode } from "enclave-manager-sdk/build/engine_service_pb";
import { useState } from "react";
import { FiEdit2 } from "react-icons/fi";
import { ArgumentValueType, KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { EnclaveFullInfo } from "../../emui/enclaves/types";
import { assertDefined, isDefined } from "../../utils";
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

  if (enclave.starlarkRun.isErr) {
    return (
      <Tooltip label={"Cannot find previous run config to edit"}>
        <Button disabled={true} colorScheme={"blue"} leftIcon={<FiEdit2 />} size={"md"}>
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
        <EditEnclaveModal
          enclave={enclave}
          kurtosisPackage={kurtosisPackage}
          onClose={() => setKurtosisPackage(undefined)}
        />
      )}
    </>
  );
};

type EditEnclaveModalProps = {
  enclave: EnclaveFullInfo;
  kurtosisPackage: KurtosisPackage;
  onClose: () => void;
};

const EditEnclaveModal = ({ enclave, kurtosisPackage, onClose }: EditEnclaveModalProps) => {
  if (enclave.starlarkRun.isErr) {
    throw Error("Internal EditEnclaveModal passed enclave with error starlark run. This should not happen.");
  }

  const parsedArgs = JSON.parse(enclave.starlarkRun.value.serializedParams);
  const convertArgValue = (
    argType: ArgumentValueType | undefined,
    value: any,
    innerType1?: ArgumentValueType,
    innerType2?: ArgumentValueType,
  ): any => {
    switch (argType) {
      case ArgumentValueType.BOOL:
        return !!value ? "true" : "false";
      case ArgumentValueType.INTEGER:
        return isDefined(value) ? `${value}` : "";
      case ArgumentValueType.STRING:
        return value || "";
      case ArgumentValueType.JSON:
        return isDefined(value) ? JSON.stringify(value) : "{}";
      case ArgumentValueType.LIST:
        assertDefined(innerType1, `Cannot parse a list argument type without knowing innerType1`);
        return isDefined(value) ? value.map((v: any) => convertArgValue(innerType1, v)) : [];
      case ArgumentValueType.DICT:
        assertDefined(innerType2, `Cannot parse a dict argument type without knowing innterType2`);
        return isDefined(value)
          ? Object.entries(value).map(([k, v]) => ({ key: k, value: convertArgValue(innerType2, v) }), {})
          : {};
      default:
        return value;
    }
  };

  const formReadyArgs = kurtosisPackage.args.reduce(
    (acc, arg) => ({
      ...acc,
      [arg.name]: convertArgValue(
        arg.typeV2?.topLevelType,
        parsedArgs[arg.name],
        arg.typeV2?.innerType1,
        arg.typeV2?.innerType2,
      ),
    }),
    {},
  );
  console.log(formReadyArgs);

  return (
    <ConfigureEnclaveModal
      isOpen={true}
      onClose={onClose}
      kurtosisPackage={kurtosisPackage}
      existingValues={{
        enclaveName: enclave.name,
        restartServices: enclave.mode === EnclaveMode.PRODUCTION,
        args: formReadyArgs,
      }}
    />
  );
};
