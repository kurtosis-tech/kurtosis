import {
  Box,
  Button,
  FormControl,
  FormLabel,
  Image,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
} from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { ChangeEvent, useState } from "react";
import {useKurtosisClient} from "../../../../../client/enclaveManager/KurtosisClientContext";

type PublishRepoModalProps = {
  isOpen: boolean;
  onClose: () => void;
  code: string;
  starlark?: string;
};

export const PublishRepoModal = ({ isOpen, onClose, code, starlark }: PublishRepoModalProps) => {
  const [repoName, setRepoName] = useState<string>("basic-package");
  const [image, setImage] = useState<File | null>(null);
  const kurtosisClient = useKurtosisClient();

  const handleImageChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setImage(e.target.files[0]);
    }
  };

  const getImageData = (file: File): Promise<Uint8Array> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = (event) => {
        const arrayBuffer = event.target?.result as ArrayBuffer;
        if (arrayBuffer) {
          const uint8Array = new Uint8Array(arrayBuffer);
          resolve(uint8Array);
        } else {
          reject(new Error('Failed to read file as ArrayBuffer.'));
        }
      };
      reader.onerror = (error) => {
        reject(error);
      };
      reader.readAsArrayBuffer(file);
    });
  };

  const handlePublishSubmit = async () => {
    console.log("Repository Name:", repoName);

    let imageData: Uint8Array = new Uint8Array(100);
    if (image) {
      imageData = await getImageData(image)
    }
    console.log("Uploaded Image:", image);

    if (!isDefined(starlark)) {
      console.log("starlark not defined which is unexpected");
      return;
    }

    const resp = await kurtosisClient.publishPackageRequest(code, repoName, starlark, imageData);
    if(resp.isOk){
      console.log(`successfully published package`)
    } else {
      console.log(`did not successfully publish package`)
    }
    return
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Enter Repository Details</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <FormControl id="repo-name" isRequired>
            <FormLabel>Package Name</FormLabel>
            <Input placeholder="Enter package name" value={repoName} onChange={(e) => setRepoName(e.target.value)} />
          </FormControl>
          <FormControl id="image-upload" mt={4}>
            <FormLabel>Upload Package Icon</FormLabel>
            <Input type="file" accept="image/*" onChange={handleImageChange} />
            {image && (
                <Box mt={4}>
                  <Image src={URL.createObjectURL(new Blob([image], { type: 'image/png' }))} alt="Uploaded image" boxSize="100px" />
                </Box>
            )}
          </FormControl>
        </ModalBody>
        <ModalFooter>
          <Button colorScheme="blue" mr={3} onClick={handlePublishSubmit}>
            Publish Package
          </Button>
          <Button onClick={onClose}>Close</Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
