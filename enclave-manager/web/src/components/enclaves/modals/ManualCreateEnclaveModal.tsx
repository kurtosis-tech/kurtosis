import {
  Button,
  Flex,
  FormControl,
  FormErrorMessage,
  FormLabel,
  Input,
  InputGroup,
  InputLeftElement,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { useState } from "react";
import { SubmitHandler, useForm } from "react-hook-form";
import { IoLogoGithub } from "react-icons/io";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useKurtosisPackageIndexerClient } from "../../../client/packageIndexer/KurtosisPackageIndexerClientContext";
import { isDefined } from "../../../utils";

type ManualCreateEnclaveForm = {
  url: string;
};

type ManualCreateEnclaveModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (kurtosisPackage: KurtosisPackage) => void;
};

export const ManualCreateEnclaveModal = ({ isOpen, onClose, onConfirm }: ManualCreateEnclaveModalProps) => {
  const kurtosisIndexerClient = useKurtosisPackageIndexerClient();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
    reset,
  } = useForm<ManualCreateEnclaveForm>();
  const [isLoading, setIsLoading] = useState(false);

  const handleClose = () => {
    reset();
    onClose();
  };

  const handleLoadSubmit: SubmitHandler<ManualCreateEnclaveForm> = async (form) => {
    setIsLoading(true);
    const packageResponse = await kurtosisIndexerClient.readPackage(form.url);
    setIsLoading(false);
    if (packageResponse.isErr) {
      setError("url", { message: `Could not load '${form.url}', got error ${packageResponse.error}` });
      return;
    }
    if (!isDefined(packageResponse.value.package)) {
      setError("url", { message: `No package found at this url` });
      return;
    }
    onConfirm(packageResponse.value.package);
    reset();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Create an Enclave</ModalHeader>
        <ModalCloseButton />
        <form onSubmit={handleSubmit(handleLoadSubmit)}>
          <ModalBody>
            <FormControl isInvalid={isDefined(errors.url)} isRequired>
              <FormLabel>Enter Github URL to package</FormLabel>
              <InputGroup>
                <InputLeftElement pointerEvents={"none"} color={"gray.400"}>
                  <IoLogoGithub />
                </InputLeftElement>
                <Input
                  {...register("url", {
                    disabled: isLoading,
                    required: true,
                  })}
                />
              </InputGroup>
              <FormErrorMessage>{errors.url?.message}</FormErrorMessage>
            </FormControl>
          </ModalBody>
          <ModalFooter>
            <Flex justifyContent={"flex-end"} gap={"12px"}>
              <Button color={"gray.100"} onClick={handleClose} disabled={isLoading}>
                Cancel
              </Button>
              <Button type={"submit"} isLoading={isLoading} colorScheme={"kurtosisGreen"}>
                Configure
              </Button>
            </Flex>
          </ModalFooter>
        </form>
      </ModalContent>
    </Modal>
  );
};
