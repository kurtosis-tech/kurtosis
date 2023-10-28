import { Button, Menu, MenuButton, MenuItem, MenuList } from "@chakra-ui/react";
import { FiPackage, FiPlus, FiSettings } from "react-icons/fi";
import { useNavigate } from "react-router-dom";

export const CreateEnclaveButton = () => {
  const navigate = useNavigate();

  return (
    <Menu matchWidth>
      <MenuButton as={Button} colorScheme={"kurtosisGreen"} leftIcon={<FiPlus />} size={"md"}>
        Create Enclave
      </MenuButton>
      <MenuList>
        <MenuItem icon={<FiSettings />}>Manual</MenuItem>
        <MenuItem onClick={() => navigate("/catalog")} icon={<FiPackage />}>
          Catalog
        </MenuItem>
      </MenuList>
    </Menu>
  );
};
