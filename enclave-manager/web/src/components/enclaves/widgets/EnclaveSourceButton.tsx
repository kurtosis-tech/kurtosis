import { Button, Icon } from "@chakra-ui/react";
import { IoLogoGithub } from "react-icons/io";

type EnclaveSourceProps = {
  source: string;
};

export const EnclaveSourceButton = ({ source }: EnclaveSourceProps) => {
  if (source.startsWith("github.com/")) {
    return (
      <Button leftIcon={<Icon as={IoLogoGithub} color={"gray.400"} />} variant={"ghost"} size={"xs"}>
        {source.replace("github.com/", "")}
      </Button>
    );
  }

  return (
    <Button variant={"ghost"} size={"xs"}>
      {source}
    </Button>
  );
};
