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
  Text,
} from "@chakra-ui/react";
import { KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import { isDefined, PackageLogo, readablePackageName } from "kurtosis-ui-components";
import { useEffect, useState } from "react";
import { FormProvider, useForm, useFormContext } from "react-hook-form";
import { CatalogContextProvider } from "../../../../catalog/CatalogContext";
import { KurtosisPackageArgumentInput } from "../../configuration/KurtosisPackageArgumentInput";
import { PackageSelector } from "../../configuration/PackageSelector";
import { transformFormArgsToKurtosisArgs, transformKurtosisArgsToFormArgs } from "../../configuration/utils";
import { KurtosisPackageNodeData } from "../types";

type ConfigurePackageNodeModalProps = {
  isOpen: boolean;
  initialValues: Record<any, string>;
  onClose: () => void;
};
export const ConfigurePackageNodeModal = ({ isOpen, onClose, initialValues }: ConfigurePackageNodeModalProps) => {
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage>();
  const parentFormMethods = useFormContext<KurtosisPackageNodeData>();
  const formMethods = useForm<Record<string, any>>();

  const onValidSubmit = (data: Record<string, any>) => {
    if (isDefined(kurtosisPackage)) {
      console.log(data);
      parentFormMethods.setValue("args", transformFormArgsToKurtosisArgs(data.args, kurtosisPackage));
      parentFormMethods.setValue("packageId", kurtosisPackage.name);
      onClose();
    }
  };

  useEffect(() => {
    if (isDefined(kurtosisPackage)) {
      console.log("setting to ", transformKurtosisArgsToFormArgs(initialValues, kurtosisPackage));
      formMethods.setValue("args", transformKurtosisArgsToFormArgs(initialValues, kurtosisPackage));
    }
  }, [kurtosisPackage, initialValues, formMethods]);

  return (
    <Modal closeOnOverlayClick={false} isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <FormProvider {...formMethods}>
        <ModalContent as={"form"} onSubmit={formMethods.handleSubmit(onValidSubmit)} h={"85vh"} minW={"800px"}>
          <ModalHeader borderBottomWidth={"1px"} borderBottomColor={"gray.500"}>
            {isDefined(kurtosisPackage) ? (
              <Flex alignItems={"center"} justifyContent={"center"} gap={"8px"}>
                <PackageLogo kurtosisPackage={kurtosisPackage} h={"24px"} display={"inline"} />
                <Text as={"span"}>{readablePackageName(kurtosisPackage.name)}</Text>
              </Flex>
            ) : (
              "Choose a package"
            )}
          </ModalHeader>
          <ModalCloseButton />
          <ModalBody overflowY={"scroll"} minH={"0"} bg={"gray.900"} pt={"32px"} pb={"32px"}>
            <CatalogContextProvider>
              {!isDefined(kurtosisPackage) && <PackageSelector onPackageSelected={setKurtosisPackage} />}
              {isDefined(kurtosisPackage) && (
                <Flex flexDirection={"column"} gap={"16px"} bg={"gray.900"}>
                  {kurtosisPackage.args.map((arg, i) => (
                    <KurtosisPackageArgumentInput key={i} argument={arg} />
                  ))}
                </Flex>
              )}
            </CatalogContextProvider>
          </ModalBody>
          <ModalFooter>
            <Flex justifyContent={"flex-end"} gap={"12px"}>
              <Button variant={"outline"} onClick={onClose}>
                Cancel
              </Button>
              <Button
                colorScheme={"kurtosisGreen"}
                variant={"outline"}
                type={"submit"}
                isDisabled={!isDefined(kurtosisPackage)}
              >
                Continue
              </Button>
            </Flex>
          </ModalFooter>
        </ModalContent>
      </FormProvider>
    </Modal>
  );
};
