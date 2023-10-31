import { Button, ButtonProps, Icon, Tag } from "@chakra-ui/react";
import { IoLogoGithub } from "react-icons/io";
import { isDefined } from "../../../utils";

type EnclaveSourceProps = ButtonProps & {
  source: string | null;
};

export const EnclaveSourceButton = ({ source, ...buttonProps }: EnclaveSourceProps) => {
  if (!isDefined(source)) {
    return <Tag>Unknown</Tag>;
  }

  if (source.startsWith("github.com/")) {
    return (
      <Button leftIcon={<Icon as={IoLogoGithub} color={"gray.400"} />} variant={"ghost"} size={"xs"} {...buttonProps}>
        {source.replace("github.com/", "")}
      </Button>
    );
  }

  return (
    <Button variant={"ghost"} size={"xs"} {...buttonProps}>
      {source}
    </Button>
  );
};
