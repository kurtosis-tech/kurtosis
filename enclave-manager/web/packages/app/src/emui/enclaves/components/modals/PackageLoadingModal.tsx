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
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined, KurtosisAlert } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { useKurtosisPackageIndexerClient } from "../../../../client/packageIndexer/KurtosisPackageIndexerClientContext";

export type PackageLoadingModalProps = {
  packageId: string;
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

const MinPackageIdLength = "github.com/".length;

export const PackageLoadingModal = ({ packageId, onPackageLoaded }: PackageLoadingModalProps) => {
  const kurtosisIndexer = useKurtosisPackageIndexerClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [isPreloading, setIsPreloading] = useState(false);
  const [loadError, setLoadError] = useState<string>();

  useEffect(() => {
    (async () => {
      if (packageId && packageId.length > MinPackageIdLength) {
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
      }
    })();
  }, [packageId, onPackageLoaded, kurtosisIndexer]);

  return (
    <Modal
      closeOnOverlayClick={false}
      isOpen={modalOpen}
      onClose={() => !isPreloading && setModalOpen(false)}
      isCentered
    >
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
            <Button color={"gray.100"} onClick={() => setModalOpen(false)} isDisabled={isPreloading}>
              Close
            </Button>
          </Flex>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
