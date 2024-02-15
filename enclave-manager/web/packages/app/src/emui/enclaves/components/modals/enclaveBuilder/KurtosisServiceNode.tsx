import { Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import { MountArtifactFileInput } from "./input/MountArtifactFileInput";
import { PortConfigurationField } from "./input/PortConfigurationInput";
import { validateDockerLocator, validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisFileMount, KurtosisPort, KurtosisServiceNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

export const KurtosisServiceNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisServiceNodeData;

    return (
      <KurtosisNode id={id} selected={selected} minWidth={650} maxWidth={800}>
        <Flex gap={"16px"}>
          <KurtosisFormControl<KurtosisServiceNodeData> name={"serviceName"} label={"Service Name"} isRequired>
            <StringArgumentInput name={"serviceName"} size={"sm"} isRequired validate={validateName} />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisServiceNodeData> name={"image"} label={"Container Image"} isRequired>
            <StringArgumentInput size={"sm"} name={"image"} isRequired validate={validateDockerLocator} />
          </KurtosisFormControl>
        </Flex>
        <Tabs>
          <TabList>
            <Tab>Environment</Tab>
            <Tab>Ports</Tab>
            <Tab>Files</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <KurtosisFormControl<KurtosisServiceNodeData> name={"env"} label={"Environment Variables"}>
                <DictArgumentInput<KurtosisServiceNodeData>
                  name={"env"}
                  KeyFieldComponent={StringArgumentInput}
                  ValueFieldComponent={MentionStringArgumentInput}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisServiceNodeData> name={"ports"} label={"Ports"}>
                <ListArgumentInput
                  name={"ports"}
                  FieldComponent={PortConfigurationField}
                  createNewValue={(): KurtosisPort => ({
                    portName: "",
                    applicationProtocol: "",
                    transportProtocol: "TCP",
                    port: 0,
                  })}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisServiceNodeData>
                name={"files"}
                label={"Files"}
                helperText={"Choose where to mount artifacts on this services filesystem"}
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
            </TabPanel>
          </TabPanels>
        </Tabs>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);
