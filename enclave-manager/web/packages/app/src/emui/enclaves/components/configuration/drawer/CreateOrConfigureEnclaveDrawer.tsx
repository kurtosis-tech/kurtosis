import { Drawer, DrawerCloseButton, DrawerContent, DrawerOverlay } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined } from "kurtosis-ui-components";
import { useEffect, useRef, useState } from "react";
import { CatalogContextProvider } from "../../../../catalog/CatalogContext";
import { EnclaveFullInfo } from "../../../types";
import { UnsavedChangesModal } from "../../modals/UnsavedChangesModal";
import { EnclaveConfigureBody, EnclaveConfigureBodyAttributes } from "./bodies/EnclaveConfigureBody";
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
  const configurationRef = useRef<EnclaveConfigureBodyAttributes>(null);
  const [drawerSize, setDrawerSize] = useState<DrawerSizes>("xl");
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage | null>(null);
  const [showConfirmCloseModal, setShowConfirmCloseModal] = useState(false);

  const handleCloseConfirmed = () => {
    setShowConfirmCloseModal(false);
    setKurtosisPackage(null);
    onClose();
  };

  const handleClose = (skipDirtyCheck?: boolean) => {
    if (skipDirtyCheck) {
      handleCloseConfirmed();
      return;
    }

    const valuesAreDirty = configurationRef.current?.isDirty() || false;

    if (valuesAreDirty) {
      setShowConfirmCloseModal(true);
    } else {
      handleCloseConfirmed();
    }
  };

  const handleToggleDrawerSize = () => {
    setDrawerSize((drawerSize) => (drawerSize === "xl" ? "full" : "xl"));
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
              ref={configurationRef}
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
      <UnsavedChangesModal
        isOpen={showConfirmCloseModal}
        onCancel={() => setShowConfirmCloseModal(false)}
        onConfirm={handleCloseConfirmed}
      />
    </Drawer>
  );
};
