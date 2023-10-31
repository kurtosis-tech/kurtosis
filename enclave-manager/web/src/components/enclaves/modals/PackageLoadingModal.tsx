import {
  Button,
  Flex,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Spinner,
  Text,
} from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useKurtosisPackageIndexerClient } from "../../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined } from "../../../utils";
import { KurtosisAlert } from "../../KurtosisAlert";

export type PackageLoadingModalProps = {
  packageId: string;
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

export const PackageLoadingModal = ({ packageId, onPackageLoaded }: PackageLoadingModalProps) => {
  const kurtosisIndexer = useKurtosisPackageIndexerClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [isPreloading, setIsPreloading] = useState(false);
  const [loadError, setLoadError] = useState<string>();

  useEffect(() => {
    (async () => {
      setModalOpen(true);
      setIsPreloading(true);
      setLoadError(undefined);
      const readPackageResponse = await kurtosisIndexer.readPackage(packageId);
      setIsPreloading(false);

      if (readPackageResponse.isErr) {
        setLoadError(readPackageResponse.error);
        return;
      }
      if (!isDefined(readPackageResponse.value.package)) {
        setLoadError(`Could not find package ${packageId}`);
        return;
      }

      setModalOpen(false);
      onPackageLoaded(readPackageResponse.value.package);
    })();
  }, [packageId, onPackageLoaded]);

  return (
    <Modal isOpen={modalOpen} onClose={() => !isPreloading && setModalOpen(false)} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Loading</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          {isPreloading && (
            <Flex flexDirection={"column"} alignItems={"center"} gap={"32px"}>
              <Spinner size={"xl"} />
              <Text>Fetching {packageId}</Text>
            </Flex>
          )}
          {isDefined(loadError) && <KurtosisAlert message={loadError} />}
        </ModalBody>
        <ModalFooter>
          <Flex justifyContent={"flex-end"} gap={"12px"}>
            <Button color={"gray.100"} onClick={() => setModalOpen(false)} disabled={isPreloading}>
              Close
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
