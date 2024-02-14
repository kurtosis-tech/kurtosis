import { Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { BooleanArgumentInput } from "../../form/BooleanArgumentInput";
import { CodeEditorInput } from "../../form/CodeEditorInput";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import { MountArtifactFileInput } from "./input/MountArtifactFileInput";
import { validateDockerLocator, validateDurationString, validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisFileMount, KurtosisShellNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

export const KurtosisShellNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisShellNodeData;

    if (!isDefined(nodeData)) {
      // Node has probably been deleted.
      return null;
    }

    return (
      <KurtosisNode
        id={id}
        selected={selected}
        name={nodeData.shellName}
        color={"purple.900"}
        minWidth={650}
        maxWidth={800}
      >
        <Flex gap={"16px"}>
          <KurtosisFormControl<KurtosisShellNodeData> name={"shellName"} label={"Shell Name"} isRequired>
            <StringArgumentInput name={"shellName"} size={"sm"} isRequired validate={validateName} />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisShellNodeData> name={"image"} label={"Container Image"}>
            <StringArgumentInput
              size={"sm"}
              name={"image"}
              validate={validateDockerLocator}
              placeholder={"Default: badouralix/curl-jq"}
            />
          </KurtosisFormControl>
        </Flex>
        <Tabs>
          <TabList>
            <Tab>Script</Tab>
            <Tab>Environment</Tab>
            <Tab>Files</Tab>
            <Tab>Advanced</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <KurtosisFormControl<KurtosisShellNodeData> name={"command"} label={"Script to run"} isRequired>
                <CodeEditorInput name={"command"} fileName={"command.sh"} isRequired />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisShellNodeData> name={"env"} label={"Environment Variables"}>
                <DictArgumentInput<KurtosisShellNodeData>
                  name={"env"}
                  KeyFieldComponent={StringArgumentInput}
                  ValueFieldComponent={MentionStringArgumentInput}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisShellNodeData>
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
              <KurtosisFormControl<KurtosisShellNodeData>
                name={"store"}
                label={"Output File/Directory"}
                helperText={
                  "Choose which files to expose from this execution task. You can use either an absolute path, a directory, or a glob."
                }
                isRequired
              >
                <MentionStringArgumentInput<KurtosisShellNodeData>
                  name={"store"}
                  placeholder={"/some/output/location"}
                  isRequired
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <Flex flexDirection={"column"} gap={"16px"}>
                <KurtosisFormControl<KurtosisShellNodeData>
                  name={"wait_enabled"}
                  label={"Wait enabled"}
                  isRequired
                  helperText={"Whether kurtosis should wait a preset time for this step to complete."}
                >
                  <BooleanArgumentInput<KurtosisShellNodeData> name={"wait_enabled"} />
                </KurtosisFormControl>
                <KurtosisFormControl<KurtosisShellNodeData>
                  name={"wait"}
                  label={"Wait"}
                  isDisabled={nodeData.wait_enabled === "false"}
                  helperText={"Whether kurtosis should wait a preset time for this step to complete."}
                >
                  <StringArgumentInput<KurtosisShellNodeData>
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
  (oldProps, newProps) => oldProps.id !== newProps.id || oldProps.selected !== newProps.selected,
);
