import { Button, ButtonProps, IconButton, IconButtonProps } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import React, { memo, MouseEventHandler, useCallback, useMemo } from "react";
import { MdBookmarkAdd } from "react-icons/md";
import { isDefined } from "../../utils";
import { useSavedPackages } from "../SavedPackages";

type SaveKurtosisPackageButtonProps<IsIconButton extends boolean> = (IsIconButton extends true
  ? IconButtonProps
  : ButtonProps) & {
  kurtosisPackage: KurtosisPackage;
  isIconButton?: IsIconButton;
};

// This button is only shown if there is a SavedPackagesContext in the tree.
export const SaveKurtosisPackageButton = <IsIconButton extends boolean>({
  kurtosisPackage,
  ...buttonProps
}: SaveKurtosisPackageButtonProps<IsIconButton>) => {
  const savePackageContext = useSavedPackages();
  if (!isDefined(savePackageContext)) {
    return null;
  }
  return <SaveKurtosisPackageButtonImpl kurtosisPackage={kurtosisPackage} {...buttonProps} />;
};

const SaveKurtosisPackageButtonImpl = <IsIconButton extends boolean>({
  kurtosisPackage,
  ...buttonProps
}: SaveKurtosisPackageButtonProps<IsIconButton>) => {
  const { savedPackages, togglePackageSaved } = useSavedPackages();
  const isPackageSaved = useMemo(
    () => savedPackages.some((p) => p.name === kurtosisPackage.name),
    [savedPackages, kurtosisPackage],
  );

  const handleClick = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      e.preventDefault();
      togglePackageSaved(kurtosisPackage);
    },
    [togglePackageSaved, kurtosisPackage],
  );

  return <SaveKurtosisPackageButtonMemo isPackageSaved={isPackageSaved} onClick={handleClick} {...buttonProps} />;
};

type SaveKurtosisPackageButtonMemoProps<IsIconButton extends boolean> = (IsIconButton extends true
  ? IconButtonProps
  : ButtonProps) & {
  isPackageSaved: boolean;
  isIconButton?: IsIconButton;

  onClick: MouseEventHandler;
};

// this is memo'd to skip unecessary renders, which effectively doubles the performance of this component (as it is
// displayed a lot.
const SaveKurtosisPackageButtonMemo = memo(
  <IsIconButton extends boolean>({
    isPackageSaved,
    onClick,
    isIconButton,
    ...buttonProps
  }: SaveKurtosisPackageButtonMemoProps<IsIconButton>) => {
    if (isIconButton) {
      return (
        <IconButton
          icon={<MdBookmarkAdd />}
          size={"xs"}
          variant={"solid"}
          colorScheme={isPackageSaved ? "kurtosisGreen" : "darkBlue"}
          onClick={onClick}
          {...(buttonProps as IconButtonProps)}
        >
          {isPackageSaved ? "Saved" : "Save"}
        </IconButton>
      );
    } else {
      return (
        <Button
          leftIcon={<MdBookmarkAdd />}
          size={"xs"}
          variant={"savedSolid"}
          colorScheme={isPackageSaved ? "kurtosisGreen" : "darkBlue"}
          bg={isPackageSaved ? "#18371E" : undefined}
          onClick={onClick}
          {...buttonProps}
        >
          {isPackageSaved ? "Saved" : "Save"}
        </Button>
      );
    }
  },
);
