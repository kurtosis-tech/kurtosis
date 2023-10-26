import { Button } from "@chakra-ui/react";
import { Link } from "react-router-dom";

type EnclaveLinkButtonProps = {
  name: string;
  uuid: string;
};

export const EnclaveLinkButton = ({ name, uuid }: EnclaveLinkButtonProps) => {
  return (
    <Link to={`/enclave/${uuid}/overview`}>
      <Button size={"sm"} variant={"ghost"}>
        {name}
      </Button>
    </Link>
  );
};
