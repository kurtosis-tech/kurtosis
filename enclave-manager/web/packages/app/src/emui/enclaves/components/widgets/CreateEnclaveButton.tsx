import { Button, Menu, MenuButton, Tooltip } from "@chakra-ui/react";
import { FiPlus } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { KURTOSIS_CREATE_ENCLAVE_URL_ARG } from "../configuration/drawer/constants";

export const CreateEnclaveButton = () => {
  const navigate = useNavigate();
  return (
    <>
      <Menu matchWidth>
        <Tooltip label={"Create a new enclave"} openDelay={1000}>
          <MenuButton
            as={Button}
            colorScheme={"green"}
            leftIcon={<FiPlus />}
            size={"sm"}
            onClick={() => navigate(`#${KURTOSIS_CREATE_ENCLAVE_URL_ARG}`)}
          >
            New Enclave
          </MenuButton>
        </Tooltip>
        {/*<MenuList>*/}
        {/*  <MenuItem onClick={() => navigate(`#${KURTOSIS_CREATE_ENCLAVE_URL_ARG}`)} icon={<FiSettings />}>*/}
        {/*    Manual*/}
        {/*  </MenuItem>*/}
        {/*  <MenuItem onClick={() => navigate("/catalog")} icon={<FiPackage />}>*/}
        {/*    Catalog*/}
        {/*  </MenuItem>*/}
        {/*</MenuList>*/}
      </Menu>
    </>
  );
};
