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
import { KurtosisPackage } from "../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useKurtosisPackageIndexerClient } from "../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined } from "../../utils";
import { KurtosisAlert } from "../KurtosisAlert";

type PreloadEnclaveProps = {
  onPackageLoaded: (kurtosisPackage: KurtosisPackage) => void;
};

export const PreloadEnclave = ({ onPackageLoaded }: PreloadEnclaveProps) => {
  const kurtosisIndexer = useKurtosisPackageIndexerClient();
  const [modalOpen, setModalOpen] = useState(false);
  const [isPreloading, setIsPreloading] = useState(false);
  const [preloadError, setPreloadError] = useState<string>();

  const searchParams = new URLSearchParams(window.location.search);
  const preloadPackage = searchParams.get("preloadPackage");

  useEffect(() => {
    (async () => {
      if (isDefined(preloadPackage)) {
        setModalOpen(true);
        setIsPreloading(true);
        setPreloadError(undefined);
        const readPackageResponse = await kurtosisIndexer.readPackage(preloadPackage);
        setIsPreloading(false);

        if (readPackageResponse.isErr) {
          setPreloadError(readPackageResponse.error);
          return;
        }
        if (!isDefined(readPackageResponse.value.package)) {
          setPreloadError(`Could not find package ${preloadPackage}`);
          return;
        }

        setModalOpen(false);
        onPackageLoaded(readPackageResponse.value.package);
      }
    })();
  }, [preloadPackage, onPackageLoaded]);

  if (!isDefined(preloadPackage)) {
    return null;
  }

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
              <Text>Fetching {preloadPackage}</Text>
            </Flex>
          )}
          {isDefined(preloadError) && <KurtosisAlert message={preloadError} />}
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
