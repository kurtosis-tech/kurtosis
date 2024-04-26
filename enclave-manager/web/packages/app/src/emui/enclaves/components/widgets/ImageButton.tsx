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
import {useMemo, useState} from "react";
import { IoLogoDocker } from "react-icons/io5";
import { useNavigate } from "react-router-dom";
import { useEnclavesContext } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";

function getUrlForImage(image: string): string | URL | undefined {
  const [imageName] = image.split(":");
  const imageParts = imageName.split("/");
  if (imageParts.length === 1) {
    return `https://hub.docker.com/_/${imageParts[0]}`;
  }
  if (imageParts.length === 2) {
    return `https://hub.docker.com/r/${imageParts[0]}/${imageParts[1]}`;
  }
  // Currently no other registries supported
  return
}

export type SetImageModalProps = {
  isOpen: boolean;
  onClose: () => void;
  currentImage: string;
  serviceName: string;
  enclave?: RemoveFunctions<EnclaveFullInfo> | undefined;
};

export const SetImageModel = ({ isOpen, onClose, currentImage, serviceName, enclave }: SetImageModalProps) => {
  const { runStarlarkScript } = useEnclavesContext(); // Assuming this is defined elsewhere
  const [newImage, setNewImage] = useState("");
  const [error, setError] = useState<string | null>(null);
  const navigator = useNavigate(); // Assuming you're using React Router's useNavigate

  const handleSetImageSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!enclave) {
      setError("Enclave is undefined. This is unexpected.");
      return;
    }

    const starlarkRun = enclave.starlarkRun;
    if (!starlarkRun || starlarkRun.isErr) {
      setError("Error: No valid Starlark run found.");
      return;
    }

    const packageId = starlarkRun.value.packageId;

    const updateImageStarlarkScript = `
    package = import_module("${packageId}/main.star")

    def run(plan, args):
      package.run(plan, **{ })
      
      plan.set_service(name="${serviceName}", config=ServiceConfig(image="${newImage}"))`;

    try {
      const logsIterator = await runStarlarkScript(enclave, updateImageStarlarkScript, {}, false);
      navigator(`/enclave/${enclave.shortenedUuid}/logs`, { state: { logs: logsIterator } });
    } catch (error: any) {
      setError(`Error updating image: ${error.message}`);
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
                    onChange={(e) => setNewImage(e.target.value)}
                />
              </FormControl>
              <Button mt={4} colorScheme="green" type="submit">
                Update
              </Button>
            </form>
          </ModalBody>
          <ModalFooter>*Note: only service and downstream dependencies will be affected.</ModalFooter>
          {error && <div style={{ color: "red", marginTop: "10px" }}>{error}</div>}
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
  const url = useMemo(() => getUrlForImage(image), [image]);
  const [showModal, setShowModal] = useState(false);

  return (
    <>
      <Icon
          as={IoLogoDocker}
          color={"gray.400"}
          boxSize={3}
          cursor={"pointer"}
          onClick={() => window.open(url, "_blank")}
          ml={2} // Adjust margin as needed
      />
      <Button
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
