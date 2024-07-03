import { Box, Flex, Icon, IconButton, Text, useToken } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { debounce } from "lodash";
import { FC, memo, PropsWithChildren, ReactElement, useCallback, useEffect, useMemo } from "react";
import { DefaultValues, FormProvider, useForm } from "react-hook-form";
import { FiCpu, FiFile, FiPackage, FiTerminal, FiTrash } from "react-icons/fi";
import { RxCornerBottomRight } from "react-icons/rx";
import { Handle, NodeResizeControl, Position, useReactFlow, useViewport } from "reactflow";
import { KurtosisNodeData } from "../types";
import { useUIState } from "../UIStateContext";
import { useVariableContext } from "../VariableContextProvider";

const colors: Record<KurtosisNodeData["type"], string> = {
  service: "blue.900",
  artifact: "yellow.900",
  shell: "red.900",
  python: "red.900",
  exec: "red.900",
  package: "kurtosisGreen.700",
};

export const nodeIcons: Record<KurtosisNodeData["type"], FC> = {
  service: FiCpu,
  artifact: FiFile,
  shell: FiTerminal,
  python: FiTerminal,
  exec: FiTerminal,
  package: FiPackage,
};

const nodeTypeReadable: Record<KurtosisNodeData["type"], string> = {
  service: "Service",
  artifact: "Files",
  exec: "Service execution task",
  shell: "Shell execution task",
  python: "Python execution task",
  package: "Package",
};

type KurtosisNodeProps = PropsWithChildren<{
  id: string;
  selected: boolean;
  minWidth: number;
  maxWidth: number;
  // Optional element to show outside of the zoom aware behaviour
  portalContent?: ReactElement;
  backgroundColor?: string;
  onClick?: () => void;
}>;

export const KurtosisNode = memo(
  <DataType extends KurtosisNodeData>({
    backgroundColor,
    children,
    id,
    maxWidth,
    minWidth,
    portalContent,
    selected,
  }: KurtosisNodeProps) => {
    const { data } = useVariableContext();
    const nodeData = data[id] as DataType;

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNodeImpl<DataType>
        backgroundColor={backgroundColor}
        id={id}
        maxWidth={maxWidth}
        minWidth={minWidth}
        nodeData={nodeData}
        portalContent={portalContent}
        selected={selected}
      >
        {children}
      </KurtosisNodeImpl>
    );
  },
);

type KurtosisNodeImplProps<DataType extends KurtosisNodeData> = KurtosisNodeProps & { nodeData: DataType };
const KurtosisNodeImpl = <DataType extends KurtosisNodeData>({
  backgroundColor,
  children,
  id,
  maxWidth,
  minWidth,
  nodeData,
  portalContent,
}: KurtosisNodeImplProps<DataType>) => {
  const { expandedNodes, toggleExpanded } = useUIState();

  const selected = Boolean(expandedNodes[id]);

  if (!selected) {
    return (
      <>
        <Handle
          type="target"
          position={Position.Left}
          style={{ left: 0, border: 0, background: "transparent" }}
          isConnectable={false}
        />
        <Handle
          type="source"
          position={Position.Right}
          style={{ right: 0, border: 0, background: "transparent" }}
          isConnectable={false}
        />
        <BasicKurtosisNode type={nodeData.type} name={nodeData.name} onClick={() => toggleExpanded(id)} />
      </>
    );
  }

  return (
    <KurtosisFormNode
      id={id}
      selected={selected}
      minWidth={minWidth}
      maxWidth={maxWidth}
      nodeData={nodeData}
      portalContent={portalContent}
      backgroundColor={backgroundColor}
      onClick={() => toggleExpanded(id)}
    >
      {children}
    </KurtosisFormNode>
  );
};

const KurtosisFormNode = <DataType extends KurtosisNodeData>({
  id,
  nodeData,
  selected,
  minWidth,
  maxWidth,
  portalContent,
  backgroundColor,
  children,
}: KurtosisNodeImplProps<DataType>) => {
  const { updateData, removeData } = useVariableContext();
  const { getNodes } = useReactFlow();
  const color = colors[nodeData.type];
  const chakraColor = useToken("colors", color);
  const formMethods = useForm<DataType>({
    defaultValues: nodeData as DefaultValues<DataType>,
    mode: "onBlur",
    shouldFocusError: false,
  });

  const { deleteElements } = useReactFlow();

  const handleDeleteNode = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();
    e.preventDefault();
    const nodesToRemove = [
      { id },
      ...getNodes()
        .filter((n) => n.parentNode === id)
        .map((n) => ({ id: n.id })),
    ];
    deleteElements({ nodes: nodesToRemove });
    removeData(nodesToRemove);
  };

  const handleChange = useMemo(
    () =>
      debounce(async () => {
        const isValid = await formMethods.trigger();
        updateData(id, (oldData) => ({ ...oldData, ...formMethods.getValues(), isValid }));
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
        boxShadow={selected ? `0 0 0 4px ${chakraColor}` : undefined}
        _hover={{ boxShadow: !selected ? `0 0 0 1px ${chakraColor}` : undefined }}
        borderColor={color}
        onBlur={handleChange}
        gap={"8px"}
      >
        <Handle
          type="target"
          position={Position.Left}
          style={{ left: 0, background: "transparent", border: "none" }}
          isConnectable={false}
        />
        <Handle
          type="source"
          position={Position.Right}
          style={{ right: 0, background: "transparent", border: "none" }}
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
        <Flex
          flexDirection={"column"}
          borderWidth={"10px"}
          borderRadius={"8px"}
          borderColor={"gray.600"}
          h={"100%"}
          w={"100%"}
          bg={backgroundColor || "gray.600"}
        >
          <ZoomAwareNodeContent
            id={id}
            name={nodeData.name}
            type={nodeData.type}
            isDisabled={nodeData.isFromPackage}
            onDelete={handleDeleteNode}
          >
            {children}
          </ZoomAwareNodeContent>
          {isDefined(portalContent) && portalContent}
        </Flex>
      </Flex>
    </FormProvider>
  );
};

type ZoomAwareNodeContentProps = PropsWithChildren<{
  name: string;
  type: KurtosisNodeData["type"];
  isDisabled?: boolean;
  id: string;
  onDelete: (e: React.MouseEvent<HTMLButtonElement>) => void;
}>;

const ZoomAwareNodeContent = ({ name, type, isDisabled, id, onDelete, children }: ZoomAwareNodeContentProps) => {
  const viewport = useViewport();
  return (
    <ZoomAwareNodeContentImpl
      name={name}
      type={type}
      isDisabled={isDisabled}
      id={id}
      onDelete={onDelete}
      zoom={viewport.zoom}
    >
      {children}
    </ZoomAwareNodeContentImpl>
  );
};

type ZoomAwareNodeContentImplProps = ZoomAwareNodeContentProps & { zoom: number };

const ZoomAwareNodeContentImpl = memo(
  ({ name, type, isDisabled, id, onDelete, zoom, children }: ZoomAwareNodeContentImplProps) => {
    const { toggleExpanded } = useUIState();
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
      return <BasicKurtosisNode name={name} type={type} />;
    }

    return (
      <>
        <Flex
          justifyContent={"space-between"}
          alignItems={"center"}
          minH={"0"}
          bg={"gray.600"}
          role="button"
          onClick={() => toggleExpanded(id)}
        >
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
            isDisabled={isDisabled}
          />
        </Flex>
        <Flex
          flexDirection={"column"}
          bg={"gray.800"}
          p={"16px 16px"}
          overflowY={"scroll"}
          className={"nodrag nowheel"}
          sx={{ cursor: "initial" }}
          onWheel={handleScroll}
          gap={"16px"}
        >
          {children}
        </Flex>
        <Box flex={"1"} w={"100%"} className={"nodrag"} />
      </>
    );
  },
);

type BasicKurtosisNodeProps = {
  type: KurtosisNodeData["type"];
  name?: string;
  onClick?: () => void;
};

const BasicKurtosisNode = ({ type, name, onClick }: BasicKurtosisNodeProps) => {
  return (
    <Flex
      gap={"20px"}
      alignItems={"center"}
      justifyContent={"center"}
      h={"100%"}
      bg={"gray.600"}
      px={"24px"}
      role="button"
      border={"2px solid"}
      borderColor={"whiteAlpha.300"}
      borderRadius={"8px"}
      onClick={onClick}
    >
      <Icon as={nodeIcons[type]} h={"40px"} w={"40px"} />
      <Text fontSize={"40px"} textAlign={"center"} p={"20px"}>
        {name || <i>Unnamed</i>}
      </Text>
    </Flex>
  );
};
