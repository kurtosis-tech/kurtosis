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
} from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { FormProvider, useForm } from "react-hook-form";
import { KurtosisFormControl } from "../../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../../form/StringArgumentInput";

type NewFileModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (newFileName: string) => void;
};
export const NewFileModal = ({ isOpen, onClose, onConfirm }: NewFileModalProps) => {
  const formMethods = useForm<{ fileName: string }>({
    defaultValues: { fileName: "" },
  });

  const onValidSubmit = (data: { fileName: string }) => {
    onConfirm(data.fileName);
  };

  return (
    <Modal closeOnOverlayClick={true} isOpen={isOpen} onClose={onClose} isCentered>
      <ModalOverlay />
      <FormProvider {...formMethods}>
        <ModalContent as={"form"} onSubmit={formMethods.handleSubmit(onValidSubmit)}>
          <ModalHeader>Create a New File</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <KurtosisFormControl
              name={"fileName"}
              label={"File Name"}
              helperText={"Enter the full file name for this file (including its path)"}
              isRequired
            >
              <StringArgumentInput
                name={"fileName"}
                placeholder={"/some/path/to/file.txt"}
                validate={(v?: string) => {
                  if (!isDefined(v)) {
                    return "input must be defined";
                  }
                  if (!v.startsWith("/")) {
                    return "File paths must start with a /";
                  }
                }}
              />
            </KurtosisFormControl>
          </ModalBody>
          <ModalFooter>
            <Flex justifyContent={"flex-end"} gap={"12px"}>
              <Button variant={"outline"} onClick={onClose}>
                Cancel
              </Button>
              <Button colorScheme={"kurtosisGreen"} variant={"outline"} type={"submit"}>
                Continue
              </Button>
            </Flex>
          </ModalFooter>
        </ModalContent>
      </FormProvider>
    </Modal>
  );
};
