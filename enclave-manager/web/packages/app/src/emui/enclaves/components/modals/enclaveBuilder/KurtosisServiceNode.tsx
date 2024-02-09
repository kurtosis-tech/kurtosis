import { Flex, Grid, GridItem, IconButton, Tab, TabList, TabPanel, TabPanels, Tabs, Text } from "@chakra-ui/react";
import { memo, useMemo } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { Handle, NodeProps, NodeResizeControl, Position, useReactFlow } from "reactflow";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { IntegerArgumentInput } from "../../form/IntegerArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { ListArgumentInput } from "../../form/ListArgumentInput";
import { OptionsArgumentInput } from "../../form/OptionArgumentInput";
import { SelectArgumentInput, SelectOption } from "../../form/SelectArgumentInput";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { MentionStringArgumentInput } from "./input/MentionStringArgumentInput";
import {
  KurtosisFileMount,
  KurtosisPort,
  KurtosisServiceNodeData,
  useVariableContext,
} from "./VariableContextProvider";

export const KurtosisServiceNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data, updateData, removeData, variables } = useVariableContext();
    const artifactVariableOptions = useMemo((): SelectOption[] => {
      return variables
        .filter((variable) => variable.id.startsWith("artifact"))
        .map((variable) => ({ display: variable.displayName, value: `{{${variable.id}}}` }));
    }, [variables]);
    const formMethods = useForm<KurtosisServiceNodeData>({
      defaultValues: (data[id] as KurtosisServiceNodeData) || {},
      mode: "onBlur",
      shouldFocusError: false,
    });

    const { deleteElements } = useReactFlow();

    const handleDeleteNode = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      e.preventDefault();
      deleteElements({ nodes: [{ id }] });
      removeData(id);
    };

    const handleBlur = async () => {
      const isValid = await formMethods.trigger();
      updateData(id, { ...formMethods.getValues(), isValid });
    };

    return (
      <FormProvider {...formMethods}>
        <Flex
          as={"form"}
          flexDirection={"column"}
          height={"100%"}
          borderRadius={"8px"}
          p={"10px"}
          bg={"gray.600"}
          borderWidth={"3px"}
          outline={"3px solid transparent"}
          outlineOffset={"3px"}
          boxShadow={selected ? "0 0 0 3px var(--chakra-colors-blue-900)" : undefined}
          borderColor={"blue.900"}
          onBlur={handleBlur}
          gap={"8px"}
        >
          <Handle
            type="target"
            position={Position.Left}
            style={{ background: "transparent", border: "none" }}
            isConnectable={false}
          />
          <Handle
            type="source"
            position={Position.Right}
            style={{ background: "transparent", border: "none" }}
            isConnectable={false}
          />
          <NodeResizeControl
            minWidth={600}
            maxWidth={800}
            minHeight={100}
            style={{ background: "transparent", border: "none" }}
          >
            <RxCornerBottomRight style={{ position: "absolute", right: 5, bottom: 5 }} />
          </NodeResizeControl>

          <Flex justifyContent={"space-between"} alignItems={"center"} minH={"0"}>
            <Text fontWeight={"semibold"}>
              {(data[id] as KurtosisServiceNodeData)?.serviceName || <i>Unnamed Service</i>}
            </Text>
            <IconButton
              aria-label={"Delete node"}
              icon={<FiTrash />}
              colorScheme={"red"}
              variant={"ghost"}
              size={"sm"}
              onClick={handleDeleteNode}
            />
          </Flex>
          <Flex
            flexDirection={"column"}
            bg={"gray.800"}
            p={"16px 16px"}
            flex={"1"}
            overflowY={"scroll"}
            className={"nodrag nowheel"}
            gap={"16px"}
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
                <StringArgumentInputsize={"sm"}
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
              </KurtosisFormControl></Flex>
            <Tabs>
              <TabList>
                <Tab>Environment Variables</Tab>
                <Tab>Ports</Tab>
                <Tab>Files</Tab>
              </TabList>

              <TabPanels>
                <TabPanel>
                  </Flex>
            <Tabs>
              <TabList>
                <Tab>Environment Variables</Tab>
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
                <TabPanel>
                  </TabPanel>
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
          </Flex>
        </Flex>
      </FormProvider>
    );
  },
  (oldProps, newProps) => {
    return oldProps.id === newProps.id && oldProps.selected === newProps.selected;
  },
);
