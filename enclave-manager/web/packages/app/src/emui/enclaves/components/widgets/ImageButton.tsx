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
import { isDefined, RemoveFunctions } from "kurtosis-ui-components";
import { useMemo, useState } from "react";
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
  return;
}

function objectToStarlark(o: any, indent: number) {
  const padLeft = "".padStart(indent, " ");
  if (!isDefined(o)) {
    return "None";
  }
  if (Array.isArray(o)) {
    let result = `[`;
    o.forEach((arrayValue) => {
      result += `${objectToStarlark(arrayValue, indent + 4)},\n`;
    });
    result += `${padLeft}],\n`;
    return result;
  }
  if (typeof o === "number") {
    return `${o}`;
  }
  if (typeof o === "string") {
    return `${o}`;
  }
  if (typeof o === "boolean") {
    return o ? "True" : "False";
  }
  if (typeof o === "object") {
    let result = "{";
    Object.entries(o).forEach(([key, value]) => {
      result += `\n${padLeft}"${key}": "${objectToStarlark(value, indent + 4)}",`;
    });
    result += `${padLeft}}`;
    return result;
  }
}

function deserializeParams(serializedParams: string): Record<string, string> {
  try {
    const parsedParams = JSON.parse(serializedParams);
    if (typeof parsedParams === "object" && parsedParams !== null) {
      const deserialized: Record<string, string> = {};
      for (const key in parsedParams) {
        if (typeof parsedParams[key] === "string") {
          deserialized[key] = parsedParams[key];
        } else {
          throw new Error("Value is not a string.");
        }
      }
      return deserialized;
    } else {
      throw new Error("Invalid JSON format.");
    }
  } catch (error) {
    console.error("Error deserializing params:", error);
    return {};
  }
}

export type SetImageModalProps = {
  isOpen: boolean;
  onClose: () => void;
  currentImage: string;
  serviceName: string;
  enclave: RemoveFunctions<EnclaveFullInfo>;
};

export const SetImageModel = ({ isOpen, onClose, currentImage, serviceName, enclave }: SetImageModalProps) => {
  const { runStarlarkScript } = useEnclavesContext(); // Assuming this is defined elsewhere
  const [error, setError] = useState<string | null>(null);
  const [newImage, setNewImage] = useState("");
  const navigator = useNavigate();

  const handleSetImageSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const starlarkRun = enclave.starlarkRun;
    if (!starlarkRun || starlarkRun.isErr) {
      setError("Error: No valid Starlark run found.");
      return;
    }

    const packageId = starlarkRun.value.packageId;

    let argsRecord;
    if (starlarkRun.value.initialSerializedParams) {
      console.log(`initial serialized params: ${starlarkRun.value.initialSerializedParams}`);
      argsRecord = deserializeParams(starlarkRun.value.initialSerializedParams);
    } else {
      console.log(`serialized params: ${starlarkRun.value.serializedParams}`);
      argsRecord = deserializeParams(starlarkRun.value.serializedParams);
    }
    const args = objectToStarlark(argsRecord, 8);
    console.log(`initial args used to start package:\n${args}`);

    const updateImageStarlarkScript = `
package = import_module("${packageId}/main.star")

def run(plan, args):
  package.run(plan, **${args})
  
  plan.set_service(name="${serviceName}", config=ServiceConfig(image="${newImage}"))`;
    console.log(`starlark script to service ${serviceName} to image ${newImage}\n${updateImageStarlarkScript}`);

    try {
      const logsIterator = await runStarlarkScript(enclave, updateImageStarlarkScript, {}, false);
      navigator(`/enclave/${enclave.shortenedUuid}/logs`, { state: { logs: logsIterator } });
    } catch (error: any) {
      setError(`Error updating image: ${error.message}`);
      console.log(error);
    }
  };

  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onClose} isCentered lockFocusAcrossFrames={undefined}>
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
  enclave: RemoveFunctions<EnclaveFullInfo>;
};

export const ImageButton = ({ image, serviceName, enclave }: ImageButtonProps) => {
  const [showModal, setShowModal] = useState(false);
  const url = useMemo(() => getUrlForImage(image), [image]);

  return (
    <>
      <Icon
        as={IoLogoDocker}
        color={"gray.400"}
        boxSize={3}
        cursor={"pointer"}
        onClick={() => window.open(url, "_blank")}
        ml={2}
      />
      <Button variant={"ghost"} size={"xs"} onClick={() => setShowModal(true)}>
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
