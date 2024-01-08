import { Button, ButtonProps, Spinner, Tag } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined, PackageLogo, readablePackageName } from "kurtosis-ui-components";
import { PropsWithChildren } from "react";
import { Link } from "react-router-dom";

type PackageLinkButtonProps = PropsWithChildren<
  ButtonProps & {
    source: "loading" | KurtosisPackage | null;
  }
>;

export const PackageLinkButton = ({ source, ...buttonProps }: PackageLinkButtonProps) => {
  if (!isDefined(source)) {
    return <Tag>Unknown</Tag>;
  }

  if (source === "loading") {
    return <Spinner size={"xs"} />;
  }

  return (
    <Link to={`/catalog/${encodeURIComponent(source.name)}`}>
      <Button variant={"ghost"} size={"xs"} leftIcon={<PackageLogo kurtosisPackage={source} w={"12px"} />}>
        {readablePackageName(source.name)}
      </Button>
    </Link>
  );
};
