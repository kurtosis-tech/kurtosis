import { Button, Menu, MenuButton, MenuItem, MenuList } from "@chakra-ui/react";
import { useState } from "react";
import { FiPackage, FiPlus, FiSettings } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../utils";
import { ConfigureEnclaveModal } from "./modals/ConfigureEnclaveModal";
import { ManualCreateEnclaveModal } from "./modals/ManualCreateEnclaveModal";

export const CreateEnclaveButton = () => {
  const navigate = useNavigate();
  const [manualCreateEnclaveOpen, setManualCreateEnclaveOpen] = useState(false);
  const [configureEnclaveOpen, setConfigureEnclaveOpen] = useState(false);
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();

  const handleManualCreateEnclaveClick = () => {
    setKurtosisPackage(undefined);
    setManualCreateEnclaveOpen(true);
  };

  const handleManualCreateEnclaveConfirmed = (kurtosisPackage: KurtosisPackage) => {
    setKurtosisPackage(kurtosisPackage);
    setManualCreateEnclaveOpen(false);
    setConfigureEnclaveOpen(true);
  };

  return (
    <>
      <Menu matchWidth>
        <MenuButton as={Button} colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
          Create Enclave
        </MenuButton>
        <MenuList>
          <MenuItem onClick={handleManualCreateEnclaveClick} icon={<FiSettings />}>
            Manual
          </MenuItem>
          <MenuItem onClick={() => navigate("/catalog")} icon={<FiPackage />}>
            Catalog
          </MenuItem>
        </MenuList>
      </Menu>
      <ManualCreateEnclaveModal
        isOpen={manualCreateEnclaveOpen}
        onClose={() => setManualCreateEnclaveOpen(false)}
        onConfirm={handleManualCreateEnclaveConfirmed}
      />
      {isDefined(kurtosisPackage) && (
        <ConfigureEnclaveModal
          isOpen={configureEnclaveOpen}
          onClose={() => setConfigureEnclaveOpen(false)}
          kurtosisPackage={kurtosisPackage}
        />
      )}
    </>
  );
};
