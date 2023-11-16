import { Button, ButtonGroup, ButtonProps, Icon, Spinner, Tag } from "@chakra-ui/react";
import { IoLogoGithub } from "react-icons/io";
import { isDefined } from "../../../utils";
import { CopyButton } from "../../CopyButton";

type EnclaveSourceProps = ButtonProps & {
  source: "loading" | string | null;
};

export const EnclaveSourceButton = ({ source, ...buttonProps }: EnclaveSourceProps) => {
  if (!isDefined(source)) {
    return <Tag>Unknown</Tag>;
  }

  if (source === "loading") {
    return <Spinner size={"xs"} />;
  }

  let button = (
    <a href={`https://${source}`} target="_blank" rel="noopener noreferrer">
      <Button variant={"ghost"} size={"xs"} {...buttonProps}>
        {source}
      </Button>
    </a>
  );
  if (source.startsWith("github.com/")) {
    button = (
      <a href={`https://${source}`} target="_blank" rel="noopener noreferrer">
        <Button leftIcon={<Icon as={IoLogoGithub} color={"gray.400"} />} variant={"ghost"} size={"xs"} {...buttonProps}>
          {source.replace("github.com/", "")}
        </Button>
      </a>
    );
  }

  return (
    <ButtonGroup>
      {button}
      <CopyButton
        contentName={"package id"}
        valueToCopy={source}
        isIconButton
        aria-label={"Copy package id"}
        size={buttonProps.size || "xs"}
      />
    </ButtonGroup>
  );
};
