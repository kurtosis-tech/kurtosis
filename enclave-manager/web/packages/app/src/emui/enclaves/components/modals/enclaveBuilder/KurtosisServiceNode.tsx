import { Flex, IconButton, Text } from "@chakra-ui/react";
import { memo } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { NodeProps, NodeResizeControl, useReactFlow } from "reactflow";
import { DictArgumentInput } from "../../form/DictArgumentInput";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";

type KurtosisPort = {
  port: number;
  transportProtocol: "TCP" | "UDP";
  applicationProtocol: string;
};

export type KurtosisServiceNodeData = {
  name: string;
  image: string;
  env: { key: string; value: string }[];
  ports: KurtosisPort[];
  isValid: boolean;
};

export const KurtosisServiceNode = memo(({ id, data, isConnectable, selected }: NodeProps<KurtosisServiceNodeData>) => {
  const formMethods = useForm<KurtosisServiceNodeData>({
    defaultValues: data,
    mode: "onBlur",
    shouldFocusError: false,
  });

  const { deleteElements, setNodes } = useReactFlow<KurtosisServiceNodeData>();

  const handleDeleteNode = () => {
    deleteElements({ nodes: [{ id }] });
  };

  const handleDataUpdate = <K extends keyof KurtosisServiceNodeData>(key: K, value: KurtosisServiceNodeData[K]) => {
    setNodes((nodes) =>
      nodes.map((node) => (node.id !== id ? node : { ...node, data: { ...node.data, [key]: value } })),
    );
  };

  const handleBlur = async () => {
    const isValid = await formMethods.trigger();
    setNodes((nodes) =>
      nodes.map((node) => (node.id !== id ? node : { ...node, data: { ...formMethods.getValues(), isValid } })),
    );
  };

  return (
    <FormProvider {...formMethods}>
      <Flex
        as={"form"}
        flexDirection={"column"}
        height={"100%"}
        borderRadius={"8px"}
        p={selected ? "8px" : "10px"}
        bg={"gray.600"}
        borderWidth={selected ? "3px" : "1px"}
        borderColor={selected ? "gray.850" : "gray.850"}
        onBlur={handleBlur}
        gap={"8px"}
      >
        {/*<Handle*/}
        {/*  type="target"*/}
        {/*  position={Position.Left}*/}
        {/*  style={{ background: "#555" }}*/}
        {/*  onConnect={(params) => console.log("handle onConnect", params)}*/}
        {/*  isConnectable={isConnectable}*/}
        {/*/>*/}
        <NodeResizeControl
          minWidth={200}
          maxWidth={600}
          minHeight={100}
          style={{ background: "transparent", border: "none" }}
        >
          <RxCornerBottomRight style={{ position: "absolute", right: 5, bottom: 5 }} />
        </NodeResizeControl>

        <Flex justifyContent={"space-between"} alignItems={"center"} minH={"0"}>
          <Text fontWeight={"semibold"}>{data.name}</Text>
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
          p={"5px 16px"}
          flex={"1"}
          overflowY={"scroll"}
          className={"nodrag nowheel"}
          gap={"8px"}
        >
          <KurtosisFormControl<KurtosisServiceNodeData> name={"name"} label={"Service Name"} isRequired>
            <StringArgumentInput
              name={"name"}
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
          <KurtosisFormControl<KurtosisServiceNodeData> name={"env"} label={"Environment Variables"}>
            <DictArgumentInput
              name={"env"}
              renderKeyFieldInput={StringArgumentInput}
              renderValueFieldInput={StringArgumentInput}
            />
          </KurtosisFormControl>
        </Flex>
      </Flex>
    </FormProvider>
  );
});
