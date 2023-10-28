import { Button, Menu, MenuButton, MenuItem, MenuList } from "@chakra-ui/react";
import { FiPackage, FiPlus, FiSettings } from "react-icons/fi";

export const CreateEnclaveButton = () => {
  return (
    <Menu matchWidth>
      <MenuButton as={Button} colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
        Create Enclave
      </MenuButton>
      <MenuList>
        <MenuItem icon={<FiSettings />}>Manual</MenuItem>
        <MenuItem icon={<FiPackage />}>Catalog</MenuItem>
      </MenuList>
    </Menu>
  );
};
