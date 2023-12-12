import { Button, ButtonGroup, ButtonProps, Icon, Link, Spinner, Tag, Tooltip } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { IoLogoGithub } from "react-icons/io";
import { useKurtosisPackageIndexerClient } from "../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined, wrapResult } from "../utils";
import { CopyButton } from "./CopyButton";

type EnclaveSourceProps = PropsWithChildren<
  ButtonProps & {
    source: "loading" | string | null;
    hideCopy?: boolean;
  }
>;

export const PackageSourceButton = ({ source, hideCopy, children, ...buttonProps }: EnclaveSourceProps) => {
  const kurtosisIndexer = useKurtosisPackageIndexerClient();

  if (!isDefined(source)) {
    return <Tag>Unknown</Tag>;
  }

  if (source === "loading") {
    return <Spinner size={"xs"} />;
  }

  let button = (
    <Link href={`https://${source}`} target="_blank" rel="noopener noreferrer" w={buttonProps.w || buttonProps.width}>
      <Button variant={"ghost"} size={"xs"} {...buttonProps}>
        {children || source}
      </Button>
    </Link>
  );
  if (source.startsWith("github.com/")) {
    const repositoryResult = wrapResult(() => kurtosisIndexer.parsePackageUrl(source));
    if (repositoryResult.isOk) {
      const repository = repositoryResult.value;
      const url = `https://${repository.baseUrl}/${repository.owner}/${repository.name}${
        isDefined(repository.rootPath) && repository.rootPath !== "/" ? "/tree/main/" + repository.rootPath : ""
      }`;

      button = (
        <Link href={url} target="_blank" rel="noopener noreferrer" w={buttonProps.w || buttonProps.width}>
          <Button
            leftIcon={<Icon as={IoLogoGithub} color={"gray.400"} />}
            variant={"ghost"}
            size={"xs"}
            {...buttonProps}
          >
            {children || source.replace("github.com/", "")}
          </Button>
        </Link>
      );
    } else {
      button = (
        <Tooltip shouldWrapChildren label={repositoryResult.error}>
          <a href={`https://${source}`} target="_blank" rel="noopener noreferrer">
            <Button variant={"ghost"} size={"xs"} {...buttonProps} colorScheme={"red"}>
              {children || source}
            </Button>
          </a>
        </Tooltip>
      );
    }
  }

  return (
    <ButtonGroup>
      {button}
      {!hideCopy && (
        <CopyButton
          contentName={"package id"}
          valueToCopy={source}
          isIconButton
          aria-label={"Copy package id"}
          size={buttonProps.size || "xs"}
        />
      )}
    </ButtonGroup>
  );
};
