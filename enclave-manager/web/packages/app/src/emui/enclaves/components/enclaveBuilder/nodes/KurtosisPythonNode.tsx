import { Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { BooleanArgumentInput } from "../../form/BooleanArgumentInput";
import { CodeEditorInput } from "../../form/CodeEditorInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../../form/types";
import { ImageConfigInput } from "../input/ImageConfigInput";
import { MentionStringArgumentInput } from "../input/MentionStringArgumentInput";
import { MountArtifactFileInput } from "../input/MountArtifactFileInput";
import { validateDurationString, validateName } from "../input/validators";
import { KurtosisFileMount, KurtosisPythonNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";
import { KurtosisNode } from "./KurtosisNode";

export const KurtosisPythonNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisPythonNodeData;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNode id={id} selected={selected} minWidth={650} maxWidth={800}>
        <Flex gap={"16px"}>
          <KurtosisFormControl<KurtosisPythonNodeData> name={"pythonName"} label={"Python Name"} isRequired>
            <StringArgumentInput name={"pythonName"} size={"sm"} isRequired validate={validateName} />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisPythonNodeData> name={"image.image"} label={"Container Image"}>
            <ImageConfigInput />
          </KurtosisFormControl>
        </Flex>
        <Tabs>
          <TabList>
            <Tab>Code</Tab>
            <Tab>Packages</Tab>
            <Tab>Arguments</Tab>
            <Tab>Files</Tab>
            <Tab>Advanced</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <KurtosisFormControl<KurtosisPythonNodeData> name={"command"} label={"Code to run"} isRequired>
                <CodeEditorInput name={"command"} fileName={`${id}.py`} isRequired />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisPythonNodeData>
                name={"packages"}
                label={"Packages"}
                isRequired
                helperText={"Names of packages that need to be installed prior to running this code"}
              >
                <ListArgumentInput<KurtosisPythonNodeData>
                  FieldComponent={PackageInput}
                  createNewValue={() => ({ packageName: "" })}
                  name={"packages"}
                  size={"sm"}
                  isRequired
                  validate={validateName}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisPythonNodeData>
                name={"args"}
                label={"Arguments"}
                helperText={"Arguments to be passed to the Python script"}
              >
                <ListArgumentInput<KurtosisPythonNodeData>
                  name={"args"}
                  FieldComponent={PythonArgInput}
                  createNewValue={() => ({ arg: "" })}
                  isRequired
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisPythonNodeData>
                name={"files"}
                label={"Input Files"}
                helperText={"Choose where to mount artifacts on this execution tasks filesystem"}
              >
                <ListArgumentInput
                  name={"files"}
                  FieldComponent={MountArtifactFileInput}
                  createNewValue={(): KurtosisFileMount => ({
                    mountPoint: "",
                    artifactName: "",
                  })}
                />
              </KurtosisFormControl>
              <KurtosisFormControl<KurtosisPythonNodeData>
                name={"store"}
                label={"Output File/Directory"}
                helperText={
                  "Choose which files to expose from this execution task. You can use either an absolute path, a directory, or a glob."
                }
              >
                <MentionStringArgumentInput<KurtosisPythonNodeData>
                  name={"store"}
                  placeholder={"/some/output/location"}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <Flex flexDirection={"column"} gap={"16px"}>
                <KurtosisFormControl<KurtosisPythonNodeData>
                  name={"wait_enabled"}
                  label={"Wait enabled"}
                  isRequired
                  helperText={"Whether kurtosis should wait a preset time for this step to complete."}
                >
                  <BooleanArgumentInput<KurtosisPythonNodeData> name={"wait_enabled"} />
                </KurtosisFormControl>
                <KurtosisFormControl<KurtosisPythonNodeData>
                  name={"wait"}
                  label={"Wait"}
                  isDisabled={nodeData.wait_enabled === "false"}
                  helperText={"Whether kurtosis should wait a preset time for this step to complete."}
                >
                  <StringArgumentInput<KurtosisPythonNodeData>
                    name={"wait"}
                    isDisabled={nodeData.wait_enabled === "false"}
                    size={"sm"}
                    placeholder={"180s"}
                    validate={nodeData.wait_enabled === "false" ? undefined : validateDurationString}
                  />
                </KurtosisFormControl>
              </Flex>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);

const PackageInput = (props: KurtosisFormInputProps<KurtosisPythonNodeData>) => {
  return (
    <StringArgumentInput<KurtosisPythonNodeData>
      {...props}
      size={"sm"}
      name={`${props.name as `packages.${number}`}.packageName`}
    />
  );
};

const PythonArgInput = (props: KurtosisFormInputProps<KurtosisPythonNodeData>) => {
  return (
    <MentionStringArgumentInput<KurtosisPythonNodeData>
      {...props}
      width={"400px"}
      name={`${props.name as `args.${number}`}.arg`}
    />
  );
};
