import { Button, ButtonGroup, ButtonProps, Icon, Spinner, Tag, Tooltip } from "@chakra-ui/react";
import { IoLogoGithub } from "react-icons/io";
import { useKurtosisPackageIndexerClient } from "../../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined, wrapResult } from "../../../utils";
import { CopyButton } from "../../CopyButton";

type EnclaveSourceProps = ButtonProps & {
  source: "loading" | string | null;
};

export const EnclaveSourceButton = ({ source, ...buttonProps }: EnclaveSourceProps) => {
  const kurtosisIndexer = useKurtosisPackageIndexerClient();

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
    const repositoryResult = wrapResult(() => kurtosisIndexer.parsePackageUrl(source));
    if (repositoryResult.isOk) {
      const repository = repositoryResult.value;
      const url = `https://${repository.baseUrl}/${repository.owner}/${repository.name}${
        isDefined(repository.rootPath) && repository.rootPath !== "/" ? "/tree/main/" + repository.rootPath : ""
      }`;

      button = (
        <a href={url} target="_blank" rel="noopener noreferrer">
          <Button
            leftIcon={<Icon as={IoLogoGithub} color={"gray.400"} />}
            variant={"ghost"}
            size={"xs"}
            {...buttonProps}
          >
            {source.replace("github.com/", "")}
          </Button>
        </a>
      );
    } else {
      button = (
        <Tooltip shouldWrapChildren label={repositoryResult.error}>
          <a href={`https://${source}`} target="_blank" rel="noopener noreferrer">
            <Button variant={"ghost"} size={"xs"} {...buttonProps} colorScheme={"red"}>
              {source}
            </Button>
          </a>
        </Tooltip>
      );
    }
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
