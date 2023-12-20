import { Image, ImageProps } from "@chakra-ui/react";
import { useHref } from "react-router-dom";
import { isDefined } from "../../utils";

type PackageLogoProps = ImageProps & {
  logoUrl?: string;
};

export const PackageLogo = ({ logoUrl, ...imageProps }: PackageLogoProps) => {
  const logoHref = useHref("/noLogo.png");
  const hasLogo = isDefined(logoUrl) && logoUrl !== "";

  return <Image src={hasLogo ? logoUrl : logoHref} fallbackSrc={logoHref} borderRadius={"6px"} {...imageProps} />;
};
