import { Button } from "@chakra-ui/react";
import { IoExitOutline } from "react-icons/io5";
import { Link } from "react-router-dom";
import { isDefined } from "../../utils";

type GotToEncalaveOverviewButtonProps = {
  enclaveUUID?: string;
};

export const GoToEnclaveOverviewButton = ({ enclaveUUID }: GotToEncalaveOverviewButtonProps) => {
  if (!isDefined(enclaveUUID)) {
    return null;
  }

  return (
    <Link to={`/enclave/${enclaveUUID}`}>
      <Button colorScheme={"kurtosisGreen"} variant={"ghost"} leftIcon={<IoExitOutline />} size={"sm"}>
        Go to Enclave Overview
      </Button>
    </Link>
  );
};
