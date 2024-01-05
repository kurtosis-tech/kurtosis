import { Drawer, DrawerCloseButton, DrawerContent, DrawerOverlay } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { CatalogContextProvider } from "../../../../catalog/CatalogContext";
import { EnclaveFullInfo } from "../../../types";
import { EnclaveConfigureBody } from "./bodies/EnclaveConfigureBody";
import { PackageSelectBody } from "./bodies/PackageSelectBody";
import { DrawerSizes } from "./types";

type CreateOrConfigureEnclaveDrawerProps = {
  isOpen: boolean;
  onClose: () => void;
  kurtosisPackage?: KurtosisPackage;
  existingEnclave?: EnclaveFullInfo;
};

export const CreateOrConfigureEnclaveDrawer = ({
  isOpen,
  onClose,
  kurtosisPackage: kurtosisPackageFromProps,
  existingEnclave,
}: CreateOrConfigureEnclaveDrawerProps) => {
  const [drawerSize, setDrawerSize] = useState<DrawerSizes>("xl");
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage | null>(null);

  const handleToggleDrawerSize = () => {
    setDrawerSize((drawerSize) => (drawerSize === "xl" ? "full" : "xl"));
  };

  const handleClose = () => {
    setKurtosisPackage(null);
    onClose();
  };

  useEffect(() => {
    if (isDefined(kurtosisPackageFromProps)) {
      setKurtosisPackage(kurtosisPackageFromProps);
    }
  }, [kurtosisPackageFromProps]);

  return (
    <Drawer isOpen={isOpen} onClose={handleClose} size={drawerSize}>
      <DrawerOverlay />
      <DrawerContent>
        <DrawerCloseButton />
        <CatalogContextProvider>
          {!isDefined(kurtosisPackage) && (
            <PackageSelectBody
              onPackageSelected={setKurtosisPackage}
              onClose={handleClose}
              drawerSize={drawerSize}
              onDrawerSizeClick={handleToggleDrawerSize}
            />
          )}
          {isDefined(kurtosisPackage) && (
            <EnclaveConfigureBody
              kurtosisPackage={kurtosisPackage}
              onBackClicked={() => setKurtosisPackage(null)}
              onClose={handleClose}
              drawerSize={drawerSize}
              onDrawerSizeClick={handleToggleDrawerSize}
              existingEnclave={existingEnclave}
            />
          )}
        </CatalogContextProvider>
      </DrawerContent>
    </Drawer>
  );
};
