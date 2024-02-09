import { Flex, IconButton, Text } from "@chakra-ui/react";
import { memo } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { Handle, NodeProps, NodeResizeControl, Position, useReactFlow } from "reactflow";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { FileTreeArgumentInput } from "./input/FileTreeArgumentInput";
import { KurtosisArtifactNodeData, useVariableContext } from "./VariableContextProvider";

export const KurtosisArtifactNode = memo(
  ({ id, selected }: NodeProps) => {
    const { data, updateData, removeData } = useVariableContext();
    const formMethods = useForm<KurtosisArtifactNodeData>({
      defaultValues: (data[id] as KurtosisArtifactNodeData) || {},
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
          boxShadow={selected ? "0 0 0 3px var(--chakra-colors-yellow-900)" : undefined}
          borderColor={"yellow.900"}
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
            minWidth={300}
            maxWidth={800}
            minHeight={100}
            style={{ background: "transparent", border: "none" }}
          >
            <RxCornerBottomRight style={{ position: "absolute", right: 5, bottom: 5 }} />
          </NodeResizeControl>

          <Flex justifyContent={"space-between"} alignItems={"center"} minH={"0"}>
            <Text fontWeight={"semibold"}>
              {(data[id] as KurtosisArtifactNodeData)?.artifactName || <i>Unnamed Artifact</i>}
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
            p={"5px 16px"}
            flex={"1"}
            overflowY={"scroll"}
            className={"nodrag nowheel"}
            gap={"8px"}
          >
            <KurtosisFormControl<KurtosisArtifactNodeData> name={"artifactName"} label={"Artifact Name"} isRequired>
              <StringArgumentInput name={"artifactName"} isRequired />
            </KurtosisFormControl>
            <KurtosisFormControl name={"files"} label={"Files"}>
              <FileTreeArgumentInput name={"files"} />
            </KurtosisFormControl>
          </Flex>
        </Flex>
      </FormProvider>
    );
  },
  (oldProps, newProps) => {
    return oldProps.id === newProps.id && oldProps.selected === newProps.selected;
  },
);
