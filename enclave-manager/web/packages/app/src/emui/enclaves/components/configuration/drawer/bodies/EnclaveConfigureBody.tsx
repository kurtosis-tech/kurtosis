import { InfoIcon } from "@chakra-ui/icons";
import {
  Box,
  Button,
  ButtonGroup,
  Card,
  CardBody,
  DrawerBody,
  DrawerFooter,
  DrawerHeader,
  Flex,
  FormControl,
  Grid,
  GridItem,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
  Tooltip,
} from "@chakra-ui/react";
import { EnclaveMode } from "enclave-manager-sdk/build/engine_service_pb";
import { ArgumentValueType, KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import {
  CopyButton,
  HoverLineTabList,
  isDefined,
  KurtosisAlert,
  KurtosisAlertError,
  PackageLogo,
  readablePackageName,
  stringifyError,
} from "kurtosis-ui-components";
import { forwardRef, useImperativeHandle, useMemo, useRef, useState } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { SubmitHandler } from "react-hook-form";
import { FiShare } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { useKurtosisClient } from "../../../../../../client/enclaveManager/KurtosisClientContext";
import { useEnclavesContext } from "../../../../EnclavesContext";
import { EnclaveFullInfo } from "../../../../types";
import { BooleanArgumentInput } from "../../../form/BooleanArgumentInput";
import { KurtosisFormControl } from "../../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../../form/StringArgumentInput";
import { allowedEnclaveNamePattern, isEnclaveNameAllowed } from "../../../utils";
import { EnclaveConfigurationForm, EnclaveConfigurationFormImperativeAttributes } from "../../EnclaveConfigurationForm";
import { KurtosisPackageArgumentInput } from "../../KurtosisPackageArgumentInput";
import { ConfigureEnclaveForm } from "../../types";
import { transformKurtosisArgsToFormArgs } from "../../utils";
import { YAMLEditorImperativeAttributes, YAMLEnclaveArgsEditor } from "../../YAMLEnclaveArgsEditor";
import { KURTOSIS_PACKAGE_ID_URL_ARG, KURTOSIS_PACKAGE_PARAMS_URL_ARG } from "../constants";
import { DrawerExpandCollapseButton } from "../DrawerExpandCollapseButton";
import { DrawerSizes } from "../types";

type EnclaveConfigureBodyProps = {
  kurtosisPackage: KurtosisPackage;
  onBackClicked: () => void;
  onClose: (skipDirtyCheck?: boolean) => void;
  drawerSize: DrawerSizes;
  onDrawerSizeClick: () => void;
  existingEnclave?: EnclaveFullInfo;
};

export type EnclaveConfigureBodyAttributes = {
  isDirty: () => boolean;
};

const tabs = ["Form", "YAML"] as const;
export const EnclaveConfigureBody = forwardRef<EnclaveConfigureBodyAttributes, EnclaveConfigureBodyProps>(
  (
    {
      kurtosisPackage,
      onBackClicked,
      onClose,
      existingEnclave,
      drawerSize,
      onDrawerSizeClick,
    }: EnclaveConfigureBodyProps,
    ref,
  ) => {
    const kurtosisClient = useKurtosisClient();
    const { createEnclave, runStarlarkPackage } = useEnclavesContext();
    const navigator = useNavigate();
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string>();
    const formRef = useRef<EnclaveConfigurationFormImperativeAttributes>(null);
    const yamlRef = useRef<YAMLEditorImperativeAttributes>(null);
    const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>("Form");

    const handleTabChange = (index: number) => {
      const newTab = tabs[index];
      if (newTab === "Form") {
        const newArgs = yamlRef.current?.getValues();
        if (!isDefined(newArgs)) {
          return;
        }
        formRef.current?.setValues("args", transformKurtosisArgsToFormArgs(newArgs, kurtosisPackage));
      }
      setActiveTab(newTab);
    };

    useImperativeHandle(
      ref,
      () => ({
        isDirty: () => {
          if (activeTab === "Form") {
            return formRef.current?.isDirty() || false;
          }
          return formRef.current?.isDirty() || yamlRef.current?.isDirty() || false;
        },
      }),
      [activeTab, formRef, yamlRef],
    );

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

          return {
            enclaveName: existingEnclave.name,
            restartServices: existingEnclave.mode === EnclaveMode.PRODUCTION,
            args: transformKurtosisArgsToFormArgs(parsedArgs, kurtosisPackage),
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
      let parsedForm: ConfigureEnclaveForm;
      try {
        parsedForm = JSON.parse(atob(preloadArgs)) as ConfigureEnclaveForm;
      } catch (err: any) {
        setError(`Unable to parse the url - was it copied correctly? Got Error: ${stringifyError(err)}`);
        return undefined;
      }
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
    }, [existingEnclave, kurtosisPackage]);

    const getLinkToCurrentConfig = () => {
      const params = new URLSearchParams({
        [KURTOSIS_PACKAGE_ID_URL_ARG]: kurtosisPackage.name,
        [KURTOSIS_PACKAGE_PARAMS_URL_ARG]: btoa(JSON.stringify(formRef.current?.getValues())),
      });

      return `${kurtosisClient.getCloudBasePathUrl()}?${params}`;
    };

    const handleLoadSubmit: SubmitHandler<ConfigureEnclaveForm> = async (formData) => {
      setError(undefined);

      let submissionData = {};

      if (activeTab === "YAML") {
        const yamlValues = yamlRef.current?.getValues();
        if (!isDefined(yamlValues)) {
          return;
        }
        formData.args = yamlValues;
      }

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

      let enclave = existingEnclave;
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
        enclave = newEnclave.value.enclaveInfo;
        enclaveUUID = newEnclave.value.enclaveInfo.shortenedUuid;
      }

      if (!isDefined(enclave)) {
        setError(`Cannot trigger starlark run as enclave info cannot be found`);
        return;
      }

      try {
        const logsIterator = await runStarlarkPackage(enclave, kurtosisPackage.name, submissionData);
        onClose(true);
        navigator(`/enclave/${enclaveUUID}/logs`, { state: { logs: logsIterator } });
      } catch (error: any) {
        setError(stringifyError(error));
      }
    };

    return (
      <ErrorBoundary fallbackRender={KurtosisAlertError}>
        <DrawerHeader as={Grid} gridTemplateColumns={"1fr 1fr 1fr"}>
          <GridItem>
            <ButtonGroup gap={"10px"}>
              <DrawerExpandCollapseButton drawerSize={drawerSize} onClick={onDrawerSizeClick} />
              <Tooltip shouldWrapChildren label={"Create a link that can be used to share this configuration."}>
                <CopyButton
                  contentName={"url"}
                  valueToCopy={getLinkToCurrentConfig}
                  aria-label={"Copy link"}
                  isIconButton
                  color={"gray.100"}
                  icon={<FiShare />}
                  size={"sm"}
                />
              </Tooltip>
            </ButtonGroup>
          </GridItem>
          <GridItem>
            <Flex alignItems={"center"} justifyContent={"center"} gap={"8px"}>
              <PackageLogo kurtosisPackage={kurtosisPackage} h={"24px"} display={"inline"} />
              <Text as={"span"}>{readablePackageName(kurtosisPackage.name)}</Text>
            </Flex>
          </GridItem>
          <GridItem />
        </DrawerHeader>
        <EnclaveConfigurationForm
          ref={formRef}
          initialValues={initialValues}
          onSubmit={handleLoadSubmit}
          kurtosisPackage={kurtosisPackage}
          style={{
            display: "flex",
            flexDirection: "column",
            flex: "1",
            minHeight: "0px",
          }}
        >
          <DrawerBody as={Flex} p={"16px"} flexDirection={"column"} gap={"16px"}>
            <Card borderWidth={"1px"} borderColor={"gray.500"}>
              <CardBody p={"0"}>
                <KurtosisFormControl name={"enclaveName"} label={"Enclave name"} type={"text"} p={"12px"}>
                  <StringArgumentInput
                    name={"enclaveName"}
                    disabled={isDefined(existingEnclave)}
                    validate={(value) => {
                      if (value.length > 0 && !isEnclaveNameAllowed(value)) {
                        return `The enclave name must match ${allowedEnclaveNamePattern}`;
                      }
                    }}
                    tabIndex={1}
                    bg={"gray.650"}
                  />
                </KurtosisFormControl>
                <FormControl
                  display={"flex"}
                  alignItems={"center"}
                  justifyContent={"space-between"}
                  gap={"16px"}
                  width={"100%"}
                  p={"12px"}
                  borderTopWidth={"1px"}
                  borderTopColor={"gray.500"}
                >
                  <Tooltip
                    shouldWrapChildren
                    label={
                      "When enabled, Kurtosis will automatically restart any services that crash inside the enclave"
                    }
                  >
                    <Text>
                      Restart services <InfoIcon ml={"6px"} w={"10px"} color={"gray.200"} />
                    </Text>
                  </Tooltip>
                  <BooleanArgumentInput inputType={"switch"} name={"restartServices"} />
                </FormControl>
              </CardBody>
            </Card>
            <Tabs onChange={handleTabChange} index={tabs.indexOf(activeTab)} isLazy>
              <HoverLineTabList tabs={tabs} activeTab={activeTab} />
              <TabPanels>
                <TabPanel>
                  <Flex flexDirection={"column"} gap={"16px"}>
                    {kurtosisPackage.args.map((arg, i) => (
                      <KurtosisPackageArgumentInput key={i} argument={arg} />
                    ))}
                  </Flex>
                </TabPanel>
                <TabPanel>
                  <YAMLEnclaveArgsEditor
                    ref={yamlRef}
                    kurtosisPackage={kurtosisPackage}
                    values={formRef.current?.getValues().args}
                    onError={setError}
                  />
                </TabPanel>
              </TabPanels>
            </Tabs>
            {isDefined(error) && (
              <Box position={"sticky"} bottom={"-16px"}>
                <KurtosisAlert message={error} onClose={() => setError(undefined)} />
              </Box>
            )}
          </DrawerBody>
          <DrawerFooter>
            <Flex flexDirection={"row-reverse"} justifyContent={"space-between"} gap={"12px"} width={"100%"}>
              <Button type={"submit"} colorScheme={"kurtosisGreen"} isLoading={isLoading}>
                {existingEnclave ? "Update" : "Run"}
              </Button>
              {!isDefined(existingEnclave) && (
                <Button color={"gray.100"} onClick={onBackClicked} isDisabled={isLoading}>
                  Back
                </Button>
              )}
            </Flex>
          </DrawerFooter>
        </EnclaveConfigurationForm>
      </ErrorBoundary>
    );
  },
);
