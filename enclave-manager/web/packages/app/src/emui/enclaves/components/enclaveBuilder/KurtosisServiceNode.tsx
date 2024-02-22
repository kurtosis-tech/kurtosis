import { Flex, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo } from "react";
import { NodeProps } from "reactflow";
import { BooleanArgumentInput } from "../form/BooleanArgumentInput";
import { DictArgumentInput } from "../form/DictArgumentInput";
import { IntegerArgumentInput } from "../form/IntegerArgumentInput";
import { KurtosisFormControl } from "../form/KurtosisFormControl";
import { ListArgumentInput } from "../form/ListArgumentInput";
import { StringArgumentInput } from "../form/StringArgumentInput";
import { KurtosisFormInputProps } from "../form/types";
import { ImageConfigInput } from "./input/ImageConfigInput";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import { MountArtifactFileInput } from "./input/MountArtifactFileInput";
import { PortConfigurationField } from "./input/PortConfigurationInput";
import { validateName } from "./input/validators";
import { KurtosisNode } from "./KurtosisNode";
import { KurtosisFileMount, KurtosisPort, KurtosisServiceNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

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
          <KurtosisFormControl<KurtosisServiceNodeData> name={"serviceName"} label={"Service Name"} isRequired>
            <StringArgumentInput name={"serviceName"} size={"sm"} isRequired validate={validateName} />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisServiceNodeData> name={"image.image"} label={"Container Image"} isRequired>
            <ImageConfigInput />
          </KurtosisFormControl>
        </Flex>
        <Tabs>
          <TabList>
            <Tab>Environment</Tab>
            <Tab>Ports</Tab>
            <Tab>Files</Tab>
            <Tab>Exec</Tab>
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
            <TabPanel>
              <Flex flexDirection={"column"} gap={"8px"}>
                <KurtosisFormControl<KurtosisServiceNodeData>
                  name={"execStepEnabled"}
                  label={"Exec step enabled"}
                  isRequired
                  helperText={"Whether kurtosis should execute a command in this service once the service is ready."}
                >
                  <BooleanArgumentInput<KurtosisServiceNodeData> name={"execStepEnabled"} />
                </KurtosisFormControl>
                <KurtosisFormControl<KurtosisServiceNodeData>
                  name={"execStepCommand"}
                  label={"Command"}
                  isRequired={nodeData.execStepEnabled === "true"}
                  isDisabled={nodeData.execStepEnabled === "false"}
                >
                  <MentionStringArgumentInput
                    size={"sm"}
                    name={"execStepCommand"}
                    isRequired={nodeData.execStepEnabled === "true"}
                    disabled={nodeData.execStepEnabled === "false"}
                  />
                </KurtosisFormControl>
                <KurtosisFormControl<KurtosisServiceNodeData>
                  name={"execStepAcceptableCodes"}
                  label={"Acceptable Exit Codes"}
                  isDisabled={nodeData.execStepEnabled === "false"}
                  helperText={
                    "If the executed command returns a code not on this list starlark will fail. Defaults to [0]"
                  }
                >
                  <ListArgumentInput<KurtosisServiceNodeData>
                    FieldComponent={AcceptableCodeInput}
                    size={"sm"}
                    name={"execStepAcceptableCodes"}
                    createNewValue={() => ({ value: 0 })}
                    disabled={nodeData.execStepEnabled === "false"}
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

const AcceptableCodeInput = (props: KurtosisFormInputProps<KurtosisServiceNodeData>) => {
  return (
    <IntegerArgumentInput<KurtosisServiceNodeData>
      {...props}
      size={"sm"}
      name={`${props.name as `execStepAcceptableCodes.${number}`}.value`}
    />
  );
};
