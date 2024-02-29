import { Button, ButtonGroup, Tooltip } from "@chakra-ui/react";
import { FiPlus, FiTool } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { KURTOSIS_BUILD_ENCLAVE_URL_ARG, KURTOSIS_CREATE_ENCLAVE_URL_ARG } from "../configuration/drawer/constants";

export const CreateEnclaveButton = () => {
  const navigate = useNavigate();
  return (
    <ButtonGroup>
      <Tooltip label={"Build a new enclave"} openDelay={1000}>
        <Button
          colorScheme={"blue"}
          leftIcon={<FiTool />}
          size={"sm"}
          onClick={() => navigate(`#${KURTOSIS_BUILD_ENCLAVE_URL_ARG}`)}
        >
          Build Enclave
        </Button>
      </Tooltip>
      <Tooltip label={"Create a new enclave"} openDelay={1000}>
        <Button
          colorScheme={"green"}
          leftIcon={<FiPlus />}
          size={"sm"}
          onClick={() => navigate(`#${KURTOSIS_CREATE_ENCLAVE_URL_ARG}`)}
        >
          New Enclave
        </Button>
      </Tooltip>
    </ButtonGroup>
  );
};
