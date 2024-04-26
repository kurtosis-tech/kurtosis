import {
  Button,
  FormControl,
  Icon,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { isDefined, RemoveFunctions, stringifyError } from "kurtosis-ui-components";
import { useState } from "react";
import { IoLogoDocker } from "react-icons/io5";
import { useNavigate } from "react-router-dom";
import { useEnclavesContext } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";

function getUrlForImage(image: string): string | null {
  const [imageName] = image.split(":");
  const imageParts = imageName.split("/");
  if (imageParts.length === 1) {
    return `https://hub.docker.com/_/${imageParts[0]}`;
  }
  if (imageParts.length === 2) {
    return `https://hub.docker.com/r/${imageParts[0]}/${imageParts[1]}`;
  }
  // Currently no other registries supported
  return null;
}

export type SetImageModalProps = {
  isOpen: boolean;
  onClose: () => void;
  currentImage: string;
  serviceName: string;
  enclave?: RemoveFunctions<EnclaveFullInfo> | undefined;
};

export const SetImageModel = ({ isOpen, onClose, currentImage, serviceName, enclave }: SetImageModalProps) => {
  const { runStarlarkScript } = useEnclavesContext();
  const [newImage, setInputValue] = useState("");
  const [error, setError] = useState<string>();
  const navigator = useNavigate();

  if (!isDefined(enclave)) {
    return null;
  }

  const handleSetImageSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log(`in handle set image ${newImage}`);

    // const serviceName = row.original.name;
    // if (!isDefined(serviceName)) {
    //   setError("No service name set.")
    //   return null
    // }
    console.log(`service name: ${serviceName}`);

    if (!isDefined(enclave.starlarkRun) || enclave.starlarkRun.isErr) {
      setError("No starlark run!");
      return null;
    }
    // TODO: make this packageId stay for consecutive runs
    let packageId = enclave.starlarkRun.value.packageId;
    console.log(`package id: ${packageId}`);

    // TODO: get the initial args of the package
    const updateImageStarlarkScript = `
package = import_module("${packageId}/main.star")

def run(plan, args):
  package.run(plan)
  
  plan.set_service(name="${serviceName}", config=ServiceConfig(image="${newImage}"))`;
    console.log(`starlark script to update image for ${serviceName} to ${newImage}:\n ${updateImageStarlarkScript}`);

    const args: Record<string, string> = {};
    try {
      const logsIterator = await runStarlarkScript(enclave, updateImageStarlarkScript, args, false);
      navigator(`/enclave/${enclave.shortenedUuid}/logs`, { state: { logs: logsIterator } });
    } catch (error: any) {
      setError(stringifyError(error));
      console.log(error);
    }
  };

  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Set new image for {serviceName}</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <form onSubmit={handleSetImageSubmit}>
            <FormControl>
              <Input
                type="text"
                name="setimage"
                placeholder={currentImage}
                value={newImage}
                onChange={(e) => setInputValue(e.target.value)}
              />
            </FormControl>
            <Button mt={4} colorScheme="green" type="submit">
              Update
            </Button>
          </form>
        </ModalBody>
        <ModalFooter>*Note: only service and downstream dependencies will be affected.</ModalFooter>
      </ModalContent>
    </Modal>
  );
};

type ImageButtonProps = {
  image: string;
  serviceName: string;
  enclave?: RemoveFunctions<EnclaveFullInfo> | undefined;
};

export const ImageButton = ({ image, serviceName, enclave }: ImageButtonProps) => {
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <Button
        leftIcon={<Icon as={IoLogoDocker} color={"gray.400"} />}
        variant={"ghost"}
        size={"xs"}
        onClick={() => setShowModal(true)}
      >
        {image}
      </Button>
      <SetImageModel
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        currentImage={image}
        serviceName={serviceName}
        enclave={enclave}
      />
    </>
  );
};
