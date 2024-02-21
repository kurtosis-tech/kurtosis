import { Flex, Icon, IconButton, Text, useToken } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { debounce } from "lodash";
import { FC, memo, PropsWithChildren, useCallback, useEffect, useMemo } from "react";
import { DefaultValues, FormProvider, useForm } from "react-hook-form";
import { FiCpu, FiFile, FiTerminal, FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { Handle, NodeResizeControl, Position, useReactFlow, useViewport } from "reactflow";
import { KurtosisNodeData } from "./types";
import { getNodeName } from "./utils";
import { useVariableContext } from "./VariableContextProvider";

const colors: Record<KurtosisNodeData["type"], string> = {
  service: "blue.900",
  artifact: "yellow.900",
  shell: "red.900",
  python: "red.900",
};

export const nodeIcons: Record<KurtosisNodeData["type"], FC> = {
  service: FiCpu,
  artifact: FiFile,
  shell: FiTerminal,
  python: FiTerminal,
};

const nodeTypeReadable: Record<KurtosisNodeData["type"], string> = {
  service: "Service",
  artifact: "Files",
  shell: "Shell execution task",
  python: "Python execution task",
};

type KurtosisNodeProps = PropsWithChildren<{
  id: string;
  selected: boolean;
  minWidth: number;
  maxWidth: number;
}>;

export const KurtosisNode = memo(
  <DataType extends KurtosisNodeData>({ id, selected, minWidth, maxWidth, children }: KurtosisNodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as DataType;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNodeImpl<DataType>
        id={id}
        selected={selected}
        minWidth={minWidth}
        maxWidth={maxWidth}
        nodeData={nodeData}
      >
        {children}
      </KurtosisNodeImpl>
    );
  },
);

type KurtosisNodeImplProps<DataType extends KurtosisNodeData> = KurtosisNodeProps & { nodeData: DataType };
const KurtosisNodeImpl = <DataType extends KurtosisNodeData>({
  id,
  nodeData,
  selected,
  minWidth,
  maxWidth,
  children,
}: KurtosisNodeImplProps<DataType>) => {
  const { updateData, removeData } = useVariableContext();
  const color = colors[nodeData.type];
  const chakraColor = useToken("colors", color);
  const name = useMemo(() => getNodeName(nodeData), [nodeData]);
  const formMethods = useForm<DataType>({
    defaultValues: nodeData as DefaultValues<DataType>,
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

  if (!isDefined(nodeData)) {
    return null;
  }

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
        <ZoomAwareNodeContent name={name} type={nodeData.type} onDelete={handleDeleteNode}>
          {children}
        </ZoomAwareNodeContent>
      </Flex>
    </FormProvider>
  );
};

type ZoomAwareNodeContentProps = PropsWithChildren<{
  name: string;
  type: KurtosisNodeData["type"];
  onDelete: (e: React.MouseEvent<HTMLButtonElement>) => void;
}>;

const ZoomAwareNodeContent = ({ name, type, onDelete, children }: ZoomAwareNodeContentProps) => {
  const viewport = useViewport();
  return (
    <ZoomAwareNodeContentImpl name={name} type={type} onDelete={onDelete} zoom={viewport.zoom}>
      {children}
    </ZoomAwareNodeContentImpl>
  );
};

type ZoomAwareNodeContentImplProps = ZoomAwareNodeContentProps & { zoom: number };

const ZoomAwareNodeContentImpl = memo(({ name, type, onDelete, zoom, children }: ZoomAwareNodeContentImplProps) => {
  const { zoomOut, zoomIn } = useReactFlow();
  const handleScroll = useCallback(
    (e: React.WheelEvent<HTMLDivElement>) => {
      if (e.currentTarget.scrollTop === 0 && e.deltaY < 0) {
        zoomIn();
      }
      if (
        Math.abs(e.currentTarget.scrollHeight - e.currentTarget.clientHeight - e.currentTarget.scrollTop) <= 1 &&
        e.deltaY > 0
      ) {
        zoomOut();
      }
    },
    [zoomOut, zoomIn],
  );

  if (zoom < 0.4) {
    return (
      <Flex gap={"20px"} alignItems={"center"} justifyContent={"center"} h={"100%"}>
        <Icon as={nodeIcons[type]} h={"40px"} w={"40px"} />
        <Text fontSize={"40px"} textAlign={"center"} p={"20px"}>
          {name || <i>Unnamed</i>}
        </Text>
      </Flex>
    );
  }

  return (
    <>
      <Flex justifyContent={"space-between"} alignItems={"center"} minH={"0"}>
        <Flex gap={"8px"} alignItems={"center"}>
          <Icon as={nodeIcons[type]} w={"20px"} h={"20px"} />
          <Text fontWeight={"semibold"}>{name || <i>Unnamed</i>}</Text>
          <Text color={"gray.300"}>
            <i>{nodeTypeReadable[type]}</i>
          </Text>
        </Flex>
        <IconButton
          className={"nodrag"}
          aria-label={"Delete node"}
          icon={<FiTrash />}
          colorScheme={"red"}
          variant={"ghost"}
          size={"sm"}
          onClick={onDelete}
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
    </>
  );
});
