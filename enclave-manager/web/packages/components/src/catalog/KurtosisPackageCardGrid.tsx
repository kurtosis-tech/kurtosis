import { Flex } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { memo } from "react";
import { KurtosisPackageCard } from "./KurtosisPackageCard";

type KurtosisPackageCardGridProps = {
  packages: KurtosisPackage[];
  onPackageRunClicked: (kurtosisPackage: KurtosisPackage) => void;
};

export const KurtosisPackageCardGrid = memo(({ packages, onPackageRunClicked }: KurtosisPackageCardGridProps) => {
  return (
    <Flex gap={"32px"} rowGap={"32px"} flexWrap={"wrap"} justifyContent={"center"}>
      {packages.map((kurtosisPackage) => (
        <KurtosisPackageCard
          kurtosisPackage={kurtosisPackage}
          onRunClick={() => onPackageRunClicked(kurtosisPackage)}
        />
      ))}
    </Flex>
  );
});
