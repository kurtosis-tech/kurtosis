import { Grid, GridItem } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { memo } from "react";
import { KurtosisPackageCard } from "./KurtosisPackageCard";

type KurtosisPackageCardGridProps = {
  packages: KurtosisPackage[];
  onPackageClicked?: (kurtosisPackage: KurtosisPackage) => void;
  onPackageRunClicked: (kurtosisPackage: KurtosisPackage) => void;
};

export const KurtosisPackageCardGrid = memo(
  ({ packages, onPackageClicked, onPackageRunClicked }: KurtosisPackageCardGridProps) => {
    return (
      <Grid gridTemplateColumns={"1fr 1fr 1fr"} columnGap={"32px"} rowGap={"32px"}>
        {packages.map((kurtosisPackage) => (
          <GridItem
            key={kurtosisPackage.url}
            onClick={onPackageClicked ? () => onPackageClicked(kurtosisPackage) : undefined}
          >
            <KurtosisPackageCard
              kurtosisPackage={kurtosisPackage}
              onRunClick={() => onPackageRunClicked(kurtosisPackage)}
            />
          </GridItem>
        ))}
      </Grid>
    );
  },
);
