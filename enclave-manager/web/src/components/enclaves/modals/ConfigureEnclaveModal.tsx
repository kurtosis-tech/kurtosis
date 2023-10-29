import {
  Button,
  Flex,
  FormControl,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
} from "@chakra-ui/react";
import { useState } from "react";
import { SubmitHandler } from "react-hook-form";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { CopyButton } from "../../CopyButton";
import { EnclaveConfigurationForm } from "../configuration/EnclaveConfigurationForm";
import { BooleanArgumentInput } from "../configuration/inputs/BooleanArgumentInput";
import { StringArgumentInput } from "../configuration/inputs/StringArgumentInput";
import { KurtosisArgumentFormControl } from "../configuration/KurtosisArgumentFormControl";
import { KurtosisPackageArgumentInput } from "../configuration/KurtosisPackageArgumentInput";
import { ConfigureEnclaveForm } from "../configuration/types";
import { EnclaveSourceButton } from "../widgets/EnclaveSourceButton";

type ConfigureEnclaveModalProps = {
  isOpen: boolean;
  onClose: () => void;
  kurtosisPackage: KurtosisPackage;
};

export const ConfigureEnclaveModal = ({ isOpen, onClose, kurtosisPackage }: ConfigureEnclaveModalProps) => {
  const [isLoading, setIsLoading] = useState(false);

  const handleClose = () => {
    onClose();
  };

  const handleLoadSubmit: SubmitHandler<ConfigureEnclaveForm> = async (form) => {
    setIsLoading(true);
    console.log(form);
    //const packageResponse = await kurtosisIndexerClient.readPackage(form.url);
    setIsLoading(false);
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} isCentered size={"5xl"}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader textAlign={"center"}>Enclave Configuration</ModalHeader>
        <ModalCloseButton />
        <EnclaveConfigurationForm onSubmit={handleLoadSubmit} kurtosisPackage={kurtosisPackage}>
          <ModalBody p={"0px"}>
            <Flex fontSize={"sm"} justifyContent={"center"} alignItems={"center"} gap={"12px"} pb={"12px"}>
              <Text>Deploying</Text>
              <EnclaveSourceButton source={kurtosisPackage.name} size={"sm"} variant={"outline"} color={"gray.100"} />
              <Text>to</Text>
              <Input size={"sm"} placeholder={"an unamed environment"} width={"auto"} />
            </Flex>
            <Flex flexDirection={"column"} gap={"24px"} p={"12px 24px"} bg={"gray.900"}>
              <Flex justifyContent={"space-between"} alignItems={"center"}>
                <FormControl display={"flex"} alignItems={"center"} gap={"16px"}>
                  <BooleanArgumentInput inputType={"switch"} name={"restartServices"} />
                  <Text fontSize={"xs"}>Restart services</Text>
                </FormControl>
                <CopyButton valueToCopy={"some value"} />
              </Flex>
              <KurtosisArgumentFormControl name={"enclaveName"} label={"Enclave name"} type={"string"}>
                <StringArgumentInput name={"enclaveName"} />
              </KurtosisArgumentFormControl>
              {kurtosisPackage.args.map((arg, i) => (
                <KurtosisPackageArgumentInput key={i} argument={arg} />
              ))}
            </Flex>
          </ModalBody>
          <ModalFooter>
            <Flex justifyContent={"flex-end"} gap={"12px"}>
              <Button color={"gray.100"} onClick={handleClose} disabled={isLoading}>
                Cancel
              </Button>
              <Button type={"submit"} isLoading={isLoading} colorScheme={"kurtosisGreen"}>
                Run
              </Button>
            </Flex>
          </ModalFooter>
        </EnclaveConfigurationForm>
      </ModalContent>
    </Modal>
  );
};
