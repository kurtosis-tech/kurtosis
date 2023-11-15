import {
  Button,
  Flex,
  FormControl,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
  Tooltip, useToast,
} from "@chakra-ui/react";
import { EnclaveMode } from "enclave-manager-sdk/build/engine_service_pb";
import { useMemo, useRef, useState } from "react";
import { SubmitHandler } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { useKurtosisClient } from "../../../client/enclaveManager/KurtosisClientContext";
import { ArgumentValueType, KurtosisPackage } from "../../../client/packageIndexer/api/kurtosis_package_indexer_pb";
import { useEmuiAppContext } from "../../../emui/EmuiAppContext";
import { EnclaveFullInfo } from "../../../emui/enclaves/types";
import { assertDefined, isDefined, stringifyError } from "../../../utils";
import { KURTOSIS_PACKAGE_ID_URL_ARG, KURTOSIS_PACKAGE_PARAMS_URL_ARG } from "../../constants";
import { CopyButton } from "../../CopyButton";
import { KurtosisAlert } from "../../KurtosisAlert";
import {
  EnclaveConfigurationForm,
  EnclaveConfigurationFormImperativeAttributes,
} from "../configuration/EnclaveConfigurationForm";
import { BooleanArgumentInput } from "../configuration/inputs/BooleanArgumentInput";
import { StringArgumentInput } from "../configuration/inputs/StringArgumentInput";
import { KurtosisArgumentFormControl } from "../configuration/KurtosisArgumentFormControl";
import { KurtosisPackageArgumentInput } from "../configuration/KurtosisPackageArgumentInput";
import { ConfigureEnclaveForm } from "../configuration/types";
import { allowedEnclaveNamePattern, isEnclaveNameAllowed } from "../utils";
import { EnclaveSourceButton } from "../widgets/EnclaveSourceButton";

type ConfigureEnclaveModalProps = {
  isOpen: boolean;
  onClose: () => void;
  kurtosisPackage: KurtosisPackage;
  existingEnclave?: EnclaveFullInfo;
};

export const ConfigureEnclaveModal = ({
  isOpen,
  onClose,
  kurtosisPackage,
  existingEnclave,
}: ConfigureEnclaveModalProps) => {
  const kurtosisClient = useKurtosisClient();
  const { createEnclave, runStarlarkPackage } = useEmuiAppContext();
  const navigator = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>();
  const formRef = useRef<EnclaveConfigurationFormImperativeAttributes>(null);
  const toast = useToast();

  const initialValues = useMemo(() => {
    if (isDefined(existingEnclave) && isDefined(existingEnclave.starlarkRun)) {
      if (existingEnclave.starlarkRun.isErr) {
        setError(
          `Could not retrieve starlark run for previous configuration, got error: ${existingEnclave.starlarkRun.isErr}`,
        );
        return undefined;
      }
      try {
        const parsedArgs = JSON.parse(existingEnclave.starlarkRun.value.serializedParams);
        const convertArgValue = (
          argType: ArgumentValueType | undefined,
          value: any,
          innerType1?: ArgumentValueType,
          innerType2?: ArgumentValueType,
        ): any => {
          switch (argType) {
            case ArgumentValueType.BOOL:
              return !!value ? "true" : isDefined(value) ? "false" : "";
            case ArgumentValueType.INTEGER:
              return isDefined(value) ? `${value}` : "";
            case ArgumentValueType.STRING:
              return value || "";
            case ArgumentValueType.JSON:
              return isDefined(value) ? JSON.stringify(value) : "{}";
            case ArgumentValueType.LIST:
              assertDefined(innerType1, `Cannot parse a list argument type without knowing innerType1`);
              return isDefined(value) ? value.map((v: any) => convertArgValue(innerType1, v)) : [];
            case ArgumentValueType.DICT:
              assertDefined(innerType2, `Cannot parse a dict argument type without knowing innterType2`);
              return isDefined(value)
                ? Object.entries(value).map(([k, v]) => ({ key: k, value: convertArgValue(innerType2, v) }), {})
                : [];
            case ArgumentValueType.JSON:
            default:
              // By default, a typeless parameter is JSON.
              return isDefined(value) ? JSON.stringify(value) : "{}";
          }
        };

        const args = kurtosisPackage.args.reduce(
          (acc, arg) => ({
            ...acc,
            [arg.name]: convertArgValue(
              arg.typeV2?.topLevelType,
              parsedArgs[arg.name],
              arg.typeV2?.innerType1,
              arg.typeV2?.innerType2,
            ),
          }),
          {},
        );
        return {
          enclaveName: existingEnclave.name,
          restartServices: existingEnclave.mode === EnclaveMode.PRODUCTION,
          args,
        } as ConfigureEnclaveForm;
      } catch (err: any) {
        setError(`Could not reuse previous configuration, got error: ${stringifyError(err)}`);
        return undefined;
      }
    }
    const searchParams = new URLSearchParams(window.location.search);
    const preloadArgs = searchParams.get(KURTOSIS_PACKAGE_PARAMS_URL_ARG);
    if (!isDefined(preloadArgs)) {
      return undefined;
    }
    const parsedForm = JSON.parse(atob(preloadArgs)) as ConfigureEnclaveForm;
    kurtosisPackage.args
      .filter((arg) => !isDefined(arg.typeV2?.topLevelType) || arg.typeV2?.topLevelType === ArgumentValueType.JSON)
      .forEach((arg) => {
        if (parsedForm.args[arg.name]) {
          try {
            parsedForm.args[arg.name] = JSON.stringify(JSON.parse(parsedForm.args[arg.name]), undefined, 4);
          } catch (err: any) {
            console.error("err", err);
            // do nothing, the input was not valid json.
          }
        }
      });
    return parsedForm;
  }, [existingEnclave, kurtosisPackage.args]);

  const getLinkToCurrentConfig = () => {
    const params = new URLSearchParams({
      [KURTOSIS_PACKAGE_ID_URL_ARG]: kurtosisPackage.name,
      [KURTOSIS_PACKAGE_PARAMS_URL_ARG]: btoa(JSON.stringify(formRef.current?.getValues())),
    });

    return `${kurtosisClient.getCloudBasePathUrl()}?${params}`;
  };

  const handleClose = () => {
    if (!isLoading) {
      navigator("#", { replace: true });
      onClose();
    }
  };

  const handleLoadSubmit: SubmitHandler<ConfigureEnclaveForm> = async (formData) => {
    setError(undefined);

    try {
      console.debug("formData", formData);
      if (formData.args && formData.args.args) {
        formData.args.args = JSON.parse(formData.args.args);
        console.debug("successfully parsed args as proper JSON", formData.args.args);
      }
    } catch (err) {
      toast({
        title: `An error occurred while preparing data for running package. The package arguments were not proper JSON: ${stringifyError("asdas")}`,
        colorScheme: "red",
      });
      return;
    }

    let apicInfo = existingEnclave?.apiContainerInfo;
    let enclaveUUID = existingEnclave?.shortenedUuid;
    if (!isDefined(existingEnclave)) {
      setIsLoading(true);
      const newEnclave = await createEnclave(formData.enclaveName, "info", formData.restartServices);
      setIsLoading(false);

      if (newEnclave.isErr) {
        setError(`Could not create enclave, got: ${newEnclave.error}`);
        return;
      }
      if (!isDefined(newEnclave.value.enclaveInfo)) {
        setError(`Did not receive enclave info when running createEnclave`);
        return;
      }
      apicInfo = newEnclave.value.enclaveInfo.apiContainerInfo;
      enclaveUUID = newEnclave.value.enclaveInfo.shortenedUuid;
    }

    if (!isDefined(apicInfo)) {
      setError(`Cannot trigger starlark run as apic info cannot be found`);
      return;
    }

    let submissionData = {};
    if (formData.args.args) {
      const { args, ...rest } = formData.args;
      submissionData = {
        ...args,
        ...rest,
      };
      console.debug("formData has `args` field and is merged with other potential args", submissionData);
    } else {
      submissionData = {
        ...formData.args,
      };
      console.debug("formData does not have Args field", submissionData);
    }
    console.log("submissionData for runStarlarkPackage", submissionData);

    const logsIterator = await runStarlarkPackage(apicInfo, kurtosisPackage.name, submissionData);
    navigator(`/enclave/${enclaveUUID}/logs`, { state: { logs: logsIterator } });
    onClose();
  };

  return (
    <Modal
      closeOnOverlayClick={false}
      isOpen={isOpen}
      onClose={handleClose}
      isCentered
      size={"5xl"}
      scrollBehavior={"inside"}
    >
      <ModalOverlay />
      <ModalContent>
        <ModalHeader flex={"0"} textAlign={"center"}>
          {!isDefined(existingEnclave) && "New "}Enclave Configuration
        </ModalHeader>
        <ModalCloseButton />
        <EnclaveConfigurationForm
          ref={formRef}
          initialValues={initialValues}
          onSubmit={handleLoadSubmit}
          kurtosisPackage={kurtosisPackage}
          style={{
            display: "flex",
            flexDirection: "column",
            flex: "0 1 auto",
            minHeight: 0,
          }}
        >
          <ModalBody flex="0 1 auto" p={"0px"} display={"flex"} flexDirection={"column"}>
            <Flex flex={"0"} fontSize={"sm"} justifyContent={"center"} alignItems={"center"} gap={"12px"} pb={"12px"}>
              <Text>Configuring</Text>
              <EnclaveSourceButton source={kurtosisPackage.name} size={"sm"} variant={"outline"} color={"gray.100"} />
            </Flex>
            {isDefined(error) && (
              <KurtosisAlert flex={"0"} message={"Could not execute configuration"} details={error} />
            )}
            <Flex
              flex={"0 1 auto"}
              overflowY={"scroll"}
              minHeight={0}
              flexDirection={"column"}
              gap={"24px"}
              p={"12px 24px"}
              bg={"gray.900"}
            >
              <Flex justifyContent={"space-between"} alignItems={"center"}>
                <Tooltip
                  shouldWrapChildren
                  label={"When enabled, Kurtosis will automatically restart any services that crash inside the enclave"}
                >
                  <FormControl display={"flex"} alignItems={"center"} gap={"16px"}>
                    <BooleanArgumentInput inputType={"switch"} name={"restartServices"} />
                    <Text fontSize={"xs"}>Restart services</Text>
                  </FormControl>
                </Tooltip>
                <Tooltip shouldWrapChildren label={"Create a link that can be used to share this configuration."}>
                  <CopyButton contentName={"url"} valueToCopy={getLinkToCurrentConfig} text={"Copy link"} />
                </Tooltip>
              </Flex>
              <KurtosisArgumentFormControl name={"enclaveName"} label={"Enclave name"} type={"string"}>
                <StringArgumentInput
                  name={"enclaveName"}
                  disabled={isDefined(existingEnclave)}
                  validate={(value) => {
                    if (value.length > 0 && !isEnclaveNameAllowed(value)) {
                      return `The enclave name must match ${allowedEnclaveNamePattern}`;
                    }
                  }}
                />
              </KurtosisArgumentFormControl>
              {kurtosisPackage.args.map((arg, i) => (
                <KurtosisPackageArgumentInput key={i} argument={arg} />
              ))}
            </Flex>
          </ModalBody>
          <ModalFooter flex={"0"}>
            <Flex justifyContent={"flex-end"} gap={"12px"}>
              <Button color={"gray.100"} onClick={handleClose} isDisabled={isLoading}>
                Cancel
              </Button>
              <Button type={"submit"} isLoading={isLoading} colorScheme={"kurtosisGreen"}>
                {existingEnclave ? "Update" : "Run"}
              </Button>
            </Flex>
          </ModalFooter>
        </EnclaveConfigurationForm>
      </ModalContent>
    </Modal>
  );
};
