import { Flex, IconButton, Text, useToken } from "@chakra-ui/react";
import { debounce } from "lodash";
import { memo, PropsWithChildren, useEffect, useMemo } from "react";
import { DefaultValues, FormProvider, useForm } from "react-hook-form";
import { FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { Handle, NodeResizeControl, Position, useReactFlow } from "reactflow";
import { KurtosisNodeData } from "./types";
import { useVariableContext } from "./VariableContextProvider";

type KurtosisNodeProps = PropsWithChildren<{
  id: string;
  name: string;
  selected: boolean;
  minWidth: number;
  maxWidth: number;
  color: string;
}>;

export const KurtosisNode = memo(
  <DataType extends KurtosisNodeData>({
    id,
    name,
    selected,
    minWidth,
    maxWidth,
    children,
    color,
  }: KurtosisNodeProps) => {
    const chakraColor = useToken("colors", color);
    const { data, updateData, removeData } = useVariableContext();
    const formMethods = useForm<DataType>({
      defaultValues: (data[id] as DefaultValues<DataType>) || {},
      mode: "onBlur",
      shouldFocusError: false,
    });

    const { deleteElements, zoomOut, zoomIn } = useReactFlow();

    const handleDeleteNode = (e: React.MouseEvent<HTMLButtonElement>) => {
      e.stopPropagation();
      e.preventDefault();
      deleteElements({ nodes: [{ id }] });
      removeData(id);
    };

    const handleChange = useMemo(
      () =>
        debounce(async () => {
          const isValid = await formMethods.trigger();
          updateData(id, { ...formMethods.getValues(), isValid });
        }, 500),
      [updateData, formMethods, id],
    );

    useEffect(() => {
      const watcher = formMethods.watch(handleChange);
      return () => watcher.unsubscribe();
    }, [formMethods, handleChange]);

    const handleScroll = (e: React.WheelEvent<HTMLDivElement>) => {
      if (e.currentTarget.scrollTop === 0 && e.deltaY < 0) {
        zoomIn();
      }
      if (
        Math.abs(e.currentTarget.scrollHeight - e.currentTarget.clientHeight - e.currentTarget.scrollTop) <= 1 &&
        e.deltaY > 0
      ) {
        zoomOut();
      }
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
          boxShadow={selected ? `0 0 0 4px ${chakraColor}` : undefined}
          _hover={{ boxShadow: !selected ? `0 0 0 1px ${chakraColor}` : undefined }}
          borderColor={color}
          onBlur={handleChange}
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
            minWidth={minWidth}
            maxWidth={maxWidth}
            minHeight={100}
            style={{ background: "transparent", border: "none" }}
          >
            <RxCornerBottomRight style={{ position: "absolute", right: 5, bottom: 5 }} />
          </NodeResizeControl>

          <Flex justifyContent={"space-between"} alignItems={"center"} minH={"0"}>
            <Text fontWeight={"semibold"}>{name || <i>Unnamed</i>}</Text>
            <IconButton
              className={"nodrag"}
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
            sx={{ cursor: "initial" }}
            onWheel={handleScroll}
            gap={"16px"}
          >
            {children}
          </Flex>
        </Flex>
      </FormProvider>
    );
  },
);
