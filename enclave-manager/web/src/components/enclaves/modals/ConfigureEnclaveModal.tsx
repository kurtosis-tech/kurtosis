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
import { useSubmit } from "react-router-dom";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { isDefined } from "../../../utils";
import { CopyButton } from "../../CopyButton";
import { KurtosisAlert } from "../../KurtosisAlert";
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
  const kurtosisClient = useKurtosisClient();
  const submit = useSubmit();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>();

  const handleClose = () => {
    onClose();
  };

  const handleLoadSubmit: SubmitHandler<ConfigureEnclaveForm> = async (formData) => {
    setIsLoading(true);
    setError(undefined);
    const newEnclave = await kurtosisClient.createEnclave(formData.enclaveName, "info", formData.restartServices);
    setIsLoading(false);

    if (newEnclave.isErr) {
      setError(`Could not create enclave, got: ${newEnclave.error.message}`);
      return;
    }
    if (!isDefined(newEnclave.value.enclaveInfo)) {
      setError(`Did not receive enclave info when running createEnclave`);
      return;
    }
    console.log(formData);
    submit(
      { config: formData, packageId: kurtosisPackage.name, enclave: newEnclave.value.enclaveInfo.toJson() },
      {
        method: "post",
        action: `/enclave/${newEnclave.value.enclaveInfo.shortenedUuid}`,
        encType: "application/json",
      },
    );
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
              {isDefined(error) && <KurtosisAlert message={error} />}
            </Flex>
            <Flex flexDirection={"column"} gap={"24px"} p={"12px 24px"} bg={"gray.900"}>
              <Flex justifyContent={"space-between"} alignItems={"center"}>
                <FormControl display={"flex"} alignItems={"center"} gap={"16px"}>
                  <BooleanArgumentInput inputType={"switch"} name={"restartServices"} />
                  <Text fontSize={"xs"}>
                    Restart services (When enabled, Kurtosis will automatically restart any services that crash inside
                    the enclave)
                  </Text>
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
