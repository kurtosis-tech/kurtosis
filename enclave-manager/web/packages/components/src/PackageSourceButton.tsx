import { Button, ButtonGroup, ButtonProps, Icon, Link, Spinner, Tag, Tooltip } from "@chakra-ui/react";
import { PropsWithChildren } from "react";
import { IoLogoGithub } from "react-icons/io";
import { isDefined, parsePackageUrl, wrapResult } from "./utils";

type EnclaveSourceProps = PropsWithChildren<
  ButtonProps & {
    source: "loading" | string | null;
  }
>;

export const PackageSourceButton = ({ source, children, ...buttonProps }: EnclaveSourceProps) => {
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
    const repositoryResult = wrapResult(() => parsePackageUrl(source));
    if (repositoryResult.isOk) {
      const repository = repositoryResult.value;
      const url = `https://${repository.baseUrl}/${repository.owner}/${repository.name}${
        isDefined(repository.rootPath) && repository.rootPath !== "/" ? "/tree/main/" + repository.rootPath : ""
      }`;

      button = (
        <Link href={url} target="_blank" rel="noopener noreferrer" w={buttonProps.w || buttonProps.width}>
          <Button leftIcon={<Icon as={IoLogoGithub} />} variant={"ghost"} size={"xs"} {...buttonProps}>
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

  return <ButtonGroup>{button}</ButtonGroup>;
};
