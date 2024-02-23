import { Button, DrawerBody, DrawerFooter, DrawerHeader, Flex, Text } from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { PackageSelector } from "../../PackageSelector";
import { DrawerExpandCollapseButton } from "../DrawerExpandCollapseButton";
import { DrawerSizes } from "../types";

type PackageSelectBodyProps = {
  onPackageSelected: (kurtosisPackage: KurtosisPackage) => void;
  onClose: () => void;
  drawerSize: DrawerSizes;
  onDrawerSizeClick: () => void;
};
export const PackageSelectBody = ({
  onPackageSelected,
  onClose,
  drawerSize,
  onDrawerSizeClick,
}: PackageSelectBodyProps) => {
  return (
    <>
      <DrawerHeader display={"flex"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
        <DrawerExpandCollapseButton drawerSize={drawerSize} onClick={onDrawerSizeClick} />
        <Text as={"span"}>Enclave Configuration</Text>
        {/*Here to balance the space-between*/}
        <Text />
      </DrawerHeader>
      <DrawerBody>
        <PackageSelector onPackageSelected={onPackageSelected} />
      </DrawerBody>
      <DrawerFooter>
        <Flex justifyContent={"space-between"} gap={"12px"} width={"100%"}>
          <Button color={"gray.100"} onClick={onClose}>
            Cancel
          </Button>
        </Flex>
      </DrawerFooter>
    </>
  );
};
