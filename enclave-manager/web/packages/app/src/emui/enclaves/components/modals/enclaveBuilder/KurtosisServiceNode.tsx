import { Flex, Grid, GridItem, Tab, TabList, TabPanel, TabPanels, Tabs } from "@chakra-ui/react";
import { useMemo } from "react";
import { NodeProps } from "reactflow";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { IntegerArgumentInput } from "../../form/IntegerArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { OptionsArgumentInput } from "../../form/OptionArgumentInput";
import { SelectArgumentInput, SelectOption } from "../../form/SelectArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import { KurtosisNode } from "./KurtosisNode";
import {
  KurtosisFileMount,
  KurtosisPort,
  KurtosisServiceNodeData,
  useVariableContext,
} from "./VariableContextProvider";

export const KurtosisServiceNode = ({ id, selected }: NodeProps) => {
  const { data, variables } = useVariableContext();
  const artifactVariableOptions = useMemo((): SelectOption[] => {
    return variables
      .filter((variable) => variable.id.startsWith("artifact"))
      .map((variable) => ({ display: variable.displayName, value: `{{${variable.id}}}` }));
  }, [variables]);

  return (
    <KurtosisNode
      id={id}
      selected={selected}
      name={(data[id] as KurtosisServiceNodeData)?.serviceName}
      minWidth={650}
      maxWidth={800}
      color={"blue.900"}
    >
      <Flex gap={"16px"}>
        <KurtosisFormControl<KurtosisServiceNodeData> name={"serviceName"} label={"Service Name"} isRequired>
          <StringArgumentInput
            name={"serviceName"}
            size={"sm"}
            isRequired
            validate={(val) => {
              if (typeof val !== "string") {
                return "Value should be a string";
              }
              if (!val.match(/^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$/)) {
                return (
                  "Service names must adhere to the RFC 1035 standard, specifically implementing this regex and" +
                  " be 1-63 characters long: ^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$. This means the service name must " +
                  "only contain lowercase alphanumeric characters or '-', and must start with a lowercase alphabet " +
                  "and end with a lowercase alphanumeric"
                );
              }
            }}
          />
        </KurtosisFormControl>
        <KurtosisFormControl<KurtosisServiceNodeData> name={"image"} label={"Container Image"} isRequired>
          <StringArgumentInput
            size={"sm"}
            name={"image"}
            isRequired
            validate={(val) => {
              if (typeof val !== "string") {
                return "Value should be a string";
              }
              if (
                !val.match(
                  /^(?<repository>[\w.\-_]+((?::\d+|)(?=\/[a-z0-9._-]+\/[a-z0-9._-]+))|)(?:\/|)(?<image>[a-z0-9.\-_]+(?:\/[a-z0-9.\-_]+|))(:(?<tag>[\w.\-_]{1,127})|)$/gim,
                )
              ) {
                return "Value does not look like a docker image";
              }
            }}
          />
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
                FieldComponent={(props) => (
                  <Grid gridTemplateColumns={"1fr 1fr"} gridGap={"8px"} p={"8px"} bgColor={"gray.650"}>
                    <GridItem>
                      <StringArgumentInput<KurtosisServiceNodeData>
                        {...props}
                        size={"sm"}
                        placeholder={"Port Name (eg postgres)"}
                        name={`${props.name as `ports.${number}`}.portName`}
                      />
                    </GridItem>
                    <GridItem>
                      <StringArgumentInput<KurtosisServiceNodeData>
                        {...props}
                        size={"sm"}
                        placeholder={"Application Protocol (eg postgresql)"}
                        name={`${props.name as `ports.${number}`}.applicationProtocol`}
                        validate={(val) => {
                          if (typeof val !== "string") {
                            return "Value should be a string";
                          }
                          if (val.includes(" ")) {
                            return "Application protocol cannot include spaces";
                          }
                        }}
                      />
                    </GridItem>
                    <GridItem>
                      <OptionsArgumentInput<KurtosisServiceNodeData>
                        {...props}
                        options={["TCP", "UDP"]}
                        name={`${props.name as `ports.${number}`}.transportProtocol`}
                      />
                    </GridItem>
                    <GridItem>
                      <IntegerArgumentInput<KurtosisServiceNodeData>
                        {...props}
                        name={`${props.name as `ports.${number}`}.port`}
                        size={"sm"}
                      />
                    </GridItem>
                  </Grid>
                )}
                createNewValue={(): KurtosisPort => ({
                  portName: "",
                  applicationProtocol: "",
                  transportProtocol: "TCP",
                  port: 0,
                })}
              />
            </KurtosisFormControl>
          </TabPanel>
          <TabPanel></TabPanel>
          <TabPanel>
            <KurtosisFormControl<KurtosisServiceNodeData>
              name={"files"}
              label={"Files"}
              helperText={"Choose where to mount artifacts on this services filesystem"}
            >
              <ListArgumentInput
                name={"files"}
                FieldComponent={(props) => (
                  <Grid gridTemplateColumns={"1fr 1fr"} gridGap={"8px"} p={"8px"} bgColor={"gray.650"}>
                    <GridItem>
                      <StringArgumentInput<KurtosisServiceNodeData>
                        {...props}
                        size={"sm"}
                        placeholder={"/some/path"}
                        name={`${props.name as `files.${number}`}.mountPoint`}
                      />
                    </GridItem>
                    <GridItem>
                      <SelectArgumentInput<KurtosisServiceNodeData>
                        options={artifactVariableOptions}
                        {...props}
                        size={"sm"}
                        placeholder={"Select an Artifact"}
                        name={`${props.name as `files.${number}`}.artifactName`}
                      />
                    </GridItem>
                  </Grid>
                )}
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
};
