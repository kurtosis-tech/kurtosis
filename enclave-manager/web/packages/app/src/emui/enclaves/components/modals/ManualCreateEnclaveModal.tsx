import { InfoIcon, SmallCloseIcon } from "@chakra-ui/icons";
import {
  Button,
  ButtonGroup,
  Card,
  CardBody,
  Drawer,
  DrawerBody,
  DrawerCloseButton,
  DrawerContent,
  DrawerFooter,
  DrawerHeader,
  DrawerOverlay,
  Flex,
  FormControl,
  Icon,
  IconButton,
  Input,
  InputGroup,
  InputLeftElement,
  InputRightElement,
  TabPanel,
  TabPanels,
  Tabs,
  Text,
  Tooltip,
  useToast,
} from "@chakra-ui/react";
import { EnclaveMode } from "enclave-manager-sdk/build/engine_service_pb";
import { ArgumentValueType, KurtosisPackage } from "kurtosis-cloud-indexer-sdk";
import {
  assertDefined,
  CopyButton,
  FindCommand,
  HoverLineTabList,
  isDefined,
  KurtosisAlert,
  KurtosisAlertError,
  KurtosisPackageCardHorizontal,
  PackageLogo,
  readablePackageName,
  stringifyError,
  useSavedPackages,
} from "kurtosis-ui-components";
import { ChangeEvent, useMemo, useRef, useState } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { SubmitHandler } from "react-hook-form";
import { BiArrowToLeft, BiArrowToRight } from "react-icons/bi";
import { FiSearch, FiShare } from "react-icons/fi";
import { useNavigate } from "react-router-dom";
import { useKurtosisClient } from "../../../../client/enclaveManager/KurtosisClientContext";
import { CatalogContextProvider, useCatalogContext } from "../../../catalog/CatalogContext";
import { useEnclavesContext } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";
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
import { KURTOSIS_PACKAGE_ID_URL_ARG, KURTOSIS_PACKAGE_PARAMS_URL_ARG } from "./constants";

type DrawerSizes = "xl" | "full";

type ManualCreateEnclaveModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (kurtosisPackage: KurtosisPackage) => void;
};

export const ManualCreateEnclaveModal = ({ isOpen, onClose, onConfirm }: ManualCreateEnclaveModalProps) => {
  const [drawerSize, setDrawerSize] = useState<DrawerSizes>("xl");
  const [kurtosisPackage, setKurtosisPackage] = useState<KurtosisPackage | null>(null);

  const handleToggleDrawerSize = () => {
    setDrawerSize((drawerSize) => (drawerSize === "xl" ? "full" : "xl"));
  };

  const handleClose = () => {
    setKurtosisPackage(null);
    onClose();
  };

  return (
    <Drawer isOpen={isOpen} onClose={handleClose} size={drawerSize}>
      <DrawerOverlay />
      <DrawerContent>
        <DrawerCloseButton />
        <CatalogContextProvider>
          {!isDefined(kurtosisPackage) && (
            <PackageSelectBody
              onPackageSelected={setKurtosisPackage}
              onClose={handleClose}
              drawerSize={drawerSize}
              onDrawerSizeClick={handleToggleDrawerSize}
            />
          )}
          {isDefined(kurtosisPackage) && (
            <EnclaveConfigureBody
              kurtosisPackage={kurtosisPackage}
              onBackClicked={() => setKurtosisPackage(null)}
              onClose={handleClose}
              drawerSize={drawerSize}
              onDrawerSizeClick={handleToggleDrawerSize}
            />
          )}
        </CatalogContextProvider>
      </DrawerContent>
    </Drawer>
  );
};

type DrawerExpandCollapseButtonProps = {
  drawerSize: DrawerSizes;
  onClick: () => void;
};

const DrawerExpandCollapseButton = ({ drawerSize, onClick }: DrawerExpandCollapseButtonProps) => {
  return (
    <IconButton
      size={"sm"}
      icon={drawerSize === "xl" ? <BiArrowToLeft /> : <BiArrowToRight />}
      aria-label={"Expand/collapse"}
      variant={"ghost"}
      onClick={onClick}
    />
  );
};

type PackageSelectBodyProps = {
  onPackageSelected: (kurtosisPackage: KurtosisPackage) => void;
  onClose: () => void;
  drawerSize: DrawerSizes;
  onDrawerSizeClick: () => void;
};

const PackageSelectBody = ({ onPackageSelected, onClose, drawerSize, onDrawerSizeClick }: PackageSelectBodyProps) => {
  const searchRef = useRef<HTMLInputElement>(null);
  const [searchTerm, setSearchTerm] = useState("");

  const { catalog } = useCatalogContext();

  const searchResults = useMemo(
    () =>
      catalog.map((catalog) =>
        catalog.packages.filter(
          (kurtosisPackage) => kurtosisPackage.name.toLowerCase().indexOf(searchTerm.toLowerCase()) >= 0,
        ),
      ),
    [catalog, searchTerm],
  );

  const { savedPackages, togglePackageSaved } = useSavedPackages();

  const handleSearchTermChange = async (e: ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  if (searchResults.isErr) {
    return (
      <DrawerBody>
        <KurtosisAlert message={"Unable to load kurtosis packages"} details={searchResults.error} />
      </DrawerBody>
    );
  }

  return (
    <>
      <DrawerHeader display={"flex"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
        <DrawerExpandCollapseButton drawerSize={drawerSize} onClick={onDrawerSizeClick} />
        <Text as={"span"}>Enclave Configuration</Text>
        {/*Here to balance the space-between*/}
        <Text />
      </DrawerHeader>
      <DrawerBody>
        <InputGroup variant={"solid"} width={"100%"} color={"gray.150"}>
          <InputLeftElement>
            <Icon as={FiSearch} />
          </InputLeftElement>
          <Input
            ref={searchRef}
            value={searchTerm}
            bgColor={"gray.850"}
            onChange={handleSearchTermChange}
            placeholder={"Search"}
            autoFocus
          />
          <InputRightElement w={"unset"} mr={"8px"}>
            {searchTerm.length > 0 ? (
              <Button variant="ghost" size={"xs"} rightIcon={<SmallCloseIcon />} onClick={() => setSearchTerm("")}>
                Clear
              </Button>
            ) : (
              <FindCommand whiteSpace={"nowrap"} pr={"10px"} />
            )}
          </InputRightElement>
        </InputGroup>
        {(searchTerm.length > 0 || savedPackages.length === 0) && (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
              {searchTerm.length === 0 ? "All Packages" : "Search Results"}
            </Text>
            {searchResults.value.map((kurtosisPackage) => (
              <KurtosisPackageCardHorizontal
                key={kurtosisPackage.name}
                kurtosisPackage={kurtosisPackage}
                onClick={() => onPackageSelected(kurtosisPackage)}
              />
            ))}
          </Flex>
        )}
        {searchTerm.length === 0 && savedPackages.length > 0 && (
          <Flex flexDirection={"column"} gap={"10px"}>
            <Text fontWeight={"semibold"} pt={"16px"} pb={"6px"}>
              Saved
            </Text>
            {savedPackages.map((kurtosisPackage) => (
              <KurtosisPackageCardHorizontal
                key={kurtosisPackage.name}
                kurtosisPackage={kurtosisPackage}
                onClick={() => onPackageSelected(kurtosisPackage)}
              />
            ))}
          </Flex>
        )}
      </DrawerBody>
      <DrawerFooter>
        <Flex justifyContent={"flex-end"} gap={"12px"}>
          <Button color={"gray.100"} onClick={onClose}>
            Cancel
          </Button>
          <Button type={"submit"} colorScheme={"kurtosisGreen"}>
            Configure
          </Button>
        </Flex>
      </DrawerFooter>
    </>
  );
};

type EnclaveConfigureBodyProps = {
  kurtosisPackage: KurtosisPackage;
  onBackClicked: () => void;
  onClose: () => void;
  drawerSize: DrawerSizes;
  onDrawerSizeClick: () => void;
  existingEnclave?: EnclaveFullInfo;
};

const tabs = ["Form", "YAML"] as const;

const EnclaveConfigureBody = ({
  kurtosisPackage,
  onBackClicked,
  onClose,
  existingEnclave,
  drawerSize,
  onDrawerSizeClick,
}: EnclaveConfigureBodyProps) => {
  const kurtosisClient = useKurtosisClient();
  const { createEnclave, runStarlarkPackage } = useEnclavesContext();
  const navigator = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>();
  const formRef = useRef<EnclaveConfigurationFormImperativeAttributes>(null);
  const toast = useToast();
  const [activeTab, setActiveTab] = useState<(typeof tabs)[number]>("Form");

  const handleTabChange = (index: number) => {
    setActiveTab(tabs[index]);
  };

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
  }, [existingEnclave, kurtosisPackage.args]);

  const getLinkToCurrentConfig = () => {
    const params = new URLSearchParams({
      [KURTOSIS_PACKAGE_ID_URL_ARG]: kurtosisPackage.name,
      [KURTOSIS_PACKAGE_PARAMS_URL_ARG]: btoa(JSON.stringify(formRef.current?.getValues())),
    });

    return `${kurtosisClient.getCloudBasePathUrl()}?${params}`;
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
        title: `An error occurred while preparing data for running package. The package arguments were not proper JSON: ${stringifyError(
          err,
        )}`,
        colorScheme: "red",
      });
      return;
    }

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

    try {
      const logsIterator = await runStarlarkPackage(enclave, kurtosisPackage.name, submissionData);
      navigator(`/enclave/${enclaveUUID}/logs`, { state: { logs: logsIterator } });
    } catch (error: any) {
      setError(stringifyError(error));
    }
  };

  return (
    <ErrorBoundary fallbackRender={KurtosisAlertError}>
      <DrawerHeader display={"flex"} justifyContent={"space-between"} alignItems={"center"} width={"100%"}>
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
        <Flex alignItems={"center"} gap={"8px"}>
          <PackageLogo kurtosisPackage={kurtosisPackage} h={"24px"} display={"inline"} />
          <Text as={"span"}>{readablePackageName(kurtosisPackage.name)}</Text>
        </Flex>
        {/*Here to balance the space-between*/}
        <Text />
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
          {isDefined(error) && (
            <KurtosisAlert flex={"1 0 auto"} message={"Could not execute configuration"} details={error} />
          )}
          <Card borderWidth={"1px"} borderColor={"gray.500"}>
            <CardBody p={"0"}>
              <KurtosisArgumentFormControl name={"enclaveName"} label={"Enclave name"} type={"text"} p={"12px"}>
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
              </KurtosisArgumentFormControl>
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
                  label={"When enabled, Kurtosis will automatically restart any services that crash inside the enclave"}
                >
                  <Text>
                    Restart services <InfoIcon ml={"6px"} w={"10px"} color={"gray.200"} />
                  </Text>
                </Tooltip>
                <BooleanArgumentInput inputType={"switch"} name={"restartServices"} />
              </FormControl>
            </CardBody>
          </Card>
          <Tabs onChange={handleTabChange}>
            <HoverLineTabList tabs={tabs} activeTab={activeTab} />
            <TabPanels>
              <TabPanel>
                <Flex flexDirection={"column"} gap={"16px"}>
                  {kurtosisPackage.args.map((arg, i) => (
                    <KurtosisPackageArgumentInput key={i} argument={arg} />
                  ))}
                </Flex>
              </TabPanel>
              <TabPanel>YAML coming soon</TabPanel>
            </TabPanels>
          </Tabs>
        </DrawerBody>
        <DrawerFooter>
          <Flex justifyContent={"space-between"} gap={"12px"}>
            <Button color={"gray.100"} onClick={onBackClicked} isDisabled={isLoading}>
              Back
            </Button>
            <Button type={"submit"} colorScheme={"kurtosisGreen"} isLoading={isLoading}>
              {existingEnclave ? "Update" : "Run"}
            </Button>
          </Flex>
        </DrawerFooter>
      </EnclaveConfigurationForm>
    </ErrorBoundary>
  );
};
