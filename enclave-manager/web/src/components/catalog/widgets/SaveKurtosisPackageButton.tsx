import { Button, ButtonProps } from "@chakra-ui/react";
import React, { memo, MouseEventHandler, useCallback, useMemo } from "react";
import { MdBookmarkAdd } from "react-icons/md";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useCatalogContext } from "../../../emui/catalog/CatalogContext";

type SaveKurtosisPackageButtonProps = ButtonProps & {
  kurtosisPackage: KurtosisPackage;
};

export const SaveKurtosisPackageButton = ({ kurtosisPackage, ...buttonProps }: SaveKurtosisPackageButtonProps) => {
  const { savedPackages, togglePackageSaved } = useCatalogContext();
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

type SaveKurtosisPackageButtonMemoProps = Omit<SaveKurtosisPackageButtonProps, "kurtosisPackage"> & {
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
