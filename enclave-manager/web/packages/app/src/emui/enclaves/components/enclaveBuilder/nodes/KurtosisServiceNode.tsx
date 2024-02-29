import { Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { ImageConfigInput } from "../input/ImageConfigInput";
import { MentionStringArgumentInput } from "../input/MentionStringArgumentInput";
import { MountArtifactFileInput } from "../input/MountArtifactFileInput";
import { PortConfigurationField } from "../input/PortConfigurationInput";
import { validateName } from "../input/validators";
import { KurtosisFileMount, KurtosisPort, KurtosisServiceNodeData } from "../types";
import { useVariableContext } from "../VariableContextProvider";
import { KurtosisNode } from "./KurtosisNode";

export const KurtosisServiceNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as KurtosisServiceNodeData;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNode id={id} selected={selected} minWidth={650} maxWidth={800}>
        <Flex gap={"16px"}>
          <KurtosisFormControl<KurtosisServiceNodeData>
            name={"name"}
            label={"Service Name"}
            isRequired
            isDisabled={nodeData.isFromPackage}
          >
            <StringArgumentInput
              name={"name"}
              size={"sm"}
              isRequired
              validate={validateName}
              isReadOnly={nodeData.isFromPackage}
            />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisServiceNodeData>
            name={"image.image"}
            label={"Container Image"}
            isRequired
            isDisabled={nodeData.isFromPackage}
          >
            <ImageConfigInput disabled={nodeData.isFromPackage} />
          </KurtosisFormControl>
        </Flex>
        <Tabs>
          <TabList>
            <Tab>Environment</Tab>
            <Tab>Ports</Tab>
            <Tab>Files</Tab>
            <Tab>Advanced</Tab>
          </TabList>
          <TabPanels>
            <TabPanel>
              <KurtosisFormControl<KurtosisServiceNodeData>
                name={"env"}
                label={"Environment Variables"}
                isDisabled={nodeData.isFromPackage}
              >
                <DictArgumentInput<KurtosisServiceNodeData>
                  name={"env"}
                  disabled={nodeData.isFromPackage}
                  KeyFieldComponent={StringArgumentInput}
                  ValueFieldComponent={MentionStringArgumentInput}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <KurtosisFormControl<KurtosisServiceNodeData>
                name={"ports"}
                label={"Ports"}
                isDisabled={nodeData.isFromPackage}
              >
                <ListArgumentInput
                  name={"ports"}
                  FieldComponent={PortConfigurationField}
                  disabled={nodeData.isFromPackage}
                  createNewValue={(): KurtosisPort => ({
                    name: "",
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
                isDisabled={nodeData.isFromPackage}
              >
                <ListArgumentInput
                  name={"files"}
                  FieldComponent={MountArtifactFileInput}
                  disabled={nodeData.isFromPackage}
                  createNewValue={(): KurtosisFileMount => ({
                    mountPoint: "",
                    name: "",
                  })}
                />
              </KurtosisFormControl>
            </TabPanel>
            <TabPanel>
              <Flex flexDirection={"column"} gap={"16px"}>
                <KurtosisFormControl<KurtosisServiceNodeData>
                  name={"entrypoint"}
                  label={"Entrypoint"}
                  helperText={
                    "The ENTRYPOINT statement hardcoded in a container image's Dockerfile might not be suitable for your needs."
                  }
                  isDisabled={nodeData.isFromPackage}
                >
                  <StringArgumentInput name={"entrypoint"} size={"sm"} isReadOnly={nodeData.isFromPackage} />
                </KurtosisFormControl>
                <KurtosisFormControl<KurtosisServiceNodeData>
                  name={"cmd"}
                  label={"CMD"}
                  helperText={
                    "The CMD statement hardcoded in a container image's Dockerfile might not be suitable for your needs."
                  }
                  isDisabled={nodeData.isFromPackage}
                >
                  <StringArgumentInput name={"cmd"} size={"sm"} isReadOnly={nodeData.isFromPackage} />
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
