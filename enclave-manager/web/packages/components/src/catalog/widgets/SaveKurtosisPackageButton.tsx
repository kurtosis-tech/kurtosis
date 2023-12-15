import { Button, ButtonProps } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import React, { memo, MouseEventHandler, useCallback, useMemo } from "react";
import { MdBookmarkAdd } from "react-icons/md";
import { isDefined } from "../../utils";
import { useSavedPackages } from "../SavedPackages";

type SaveKurtosisPackageButtonProps = ButtonProps & {
  kurtosisPackage: KurtosisPackage;
};

// This button is only shown if there is a SavedPackagesContext in the tree.
export const SaveKurtosisPackageButton = ({ kurtosisPackage, ...buttonProps }: SaveKurtosisPackageButtonProps) => {
  const savePackageContext = useSavedPackages();
  if (!isDefined(savePackageContext)) {
    return null;
  }
  return <SaveKurtosisPackageButtonImpl kurtosisPackage={kurtosisPackage} {...buttonProps} />;
};

const SaveKurtosisPackageButtonImpl = ({ kurtosisPackage, ...buttonProps }: SaveKurtosisPackageButtonProps) => {
  const { savedPackages, togglePackageSaved } = useSavedPackages();
  const isPackageSaved = useMemo(
    () => savedPackages.some((p) => p.name === kurtosisPackage.name),
    [savedPackages, kurtosisPackage],
  );

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.preventDefault();
      togglePackageSaved(kurtosisPackage);
    },
    [togglePackageSaved, kurtosisPackage],
  );

  return <SaveKurtosisPackageButtonMemo isPackageSaved={isPackageSaved} onClick={handleClick} {...buttonProps} />;
};

type SaveKurtosisPackageButtonMemoProps = ButtonProps & {
  isPackageSaved: boolean;
  onClick: MouseEventHandler;
};

// this is memo'd to skip unecessary renders, which effectively doubles the performance of this component (as it is
// displayed a lot.
const SaveKurtosisPackageButtonMemo = memo(
  ({ isPackageSaved, onClick, ...buttonProps }: SaveKurtosisPackageButtonMemoProps) => {
    return (
      <Button
        size={"xs"}
        variant={"solid"}
        colorScheme={isPackageSaved ? "kurtosisGreen" : "darkBlue"}
        leftIcon={<MdBookmarkAdd />}
        onClick={onClick}
        bg={isPackageSaved ? "#18371E" : undefined}
        {...buttonProps}
      >
        {isPackageSaved ? "Saved" : "Save"}
      </Button>
    );
  },
);
