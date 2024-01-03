import { Image, ImageProps } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { useHref } from "react-router-dom";
import { isDefined } from "../../utils";

type PackageLogoProps = ImageProps & {
  kurtosisPackage: KurtosisPackage;
};

export const PackageLogo = ({ kurtosisPackage, ...imageProps }: PackageLogoProps) => {
  const noLogoHref = useHref("/noLogo.png");
  const kurtosisLogoHref = useHref("/logo.png");

  const hasLogo = isDefined(kurtosisPackage.iconUrl) && kurtosisPackage.iconUrl !== "";
  const isKurtosisPackage = kurtosisPackage.repositoryMetadata?.owner === "kurtosis-tech";

  return (
    <Image
      src={hasLogo ? kurtosisPackage.iconUrl : isKurtosisPackage ? kurtosisLogoHref : noLogoHref}
      fallbackSrc={noLogoHref}
      borderRadius={"6px"}
      {...imageProps}
    />
  );
};
