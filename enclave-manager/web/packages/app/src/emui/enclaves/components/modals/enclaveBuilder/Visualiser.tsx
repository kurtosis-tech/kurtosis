import { Box, Button, ButtonGroup, Flex } from "@chakra-ui/react";
import Dagre from "@dagrejs/dagre";
import { RemoveFunctions } from "kurtosis-ui-components";
import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef } from "react";
import { FiPlusCircle } from "react-icons/fi";
import {
  Background,
  BackgroundVariant,
  Controls,
  Edge,
  Node,
  ReactFlow,
  useEdgesState,
  useNodesState,
  useReactFlow,
  XYPosition,
} from "reactflow";
import { v4 as uuidv4 } from "uuid";
import { EnclaveFullInfo } from "../../../types";
import { KurtosisArtifactNode } from "./KurtosisArtifactNode";
import { KurtosisExecNode } from "./KurtosisExecNode";
import { KurtosisServiceNode } from "./KurtosisServiceNode";
import { KurtosisShellNode } from "./KurtosisShellNode";
import { generateStarlarkFromGraph, getNodeDependencies } from "./utils";
import { useVariableContext } from "./VariableContextProvider";
import "./Visualiser.css";

const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));
const getLayoutedElements = <T extends object>(nodes: Node<T>[], edges: Edge<any>[]) => {
  if (nodes.length === 0) {
    return { nodes, edges };
  }
  g.setGraph({ rankdir: "LR", ranksep: 100 });

  edges.forEach((edge) => g.setEdge(edge.source, edge.target));
  nodes.forEach((node) =>
    g.setNode(node.id, node as Node<{ label: string }, string | undefined> & { width?: number; height?: number }),
  );

  Dagre.layout(g);

  return {
    nodes: nodes.map((node) => {
      const { x, y } = g.node(node.id);

      return { ...node, position: { x, y } };
    }),
    edges,
  };
};

const nodeTypes = {
  serviceNode: KurtosisServiceNode,
  artifactNode: KurtosisArtifactNode,
  shellNode: KurtosisShellNode,
  execNode: KurtosisExecNode,
};

export type VisualiserImperativeAttributes = {
  getStarlark: () => string;
};
type VisualiserProps = {
  initialNodes: Node<any>[];
  initialEdges: Edge<any>[];
  existingEnclave?: RemoveFunctions<EnclaveFullInfo>;
};
export const Visualiser = forwardRef<VisualiserImperativeAttributes, VisualiserProps>(
  ({ initialNodes, initialEdges, existingEnclave }, ref) => {
    const { data, updateData } = useVariableContext();
    const insertOffset = useRef(0);
    const { fitView, addNodes, getViewport } = useReactFlow();
    const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes || []);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges || []);

    const onLayout = useCallback(() => {
      const layouted = getLayoutedElements(nodes, edges);

      setNodes([...layouted.nodes]);
      setEdges([...layouted.edges]);

      window.requestAnimationFrame(() => {
        fitView();
      });
    }, [nodes, edges, fitView, setEdges, setNodes]);

    const getNewNodePosition = (): XYPosition => {
      const viewport = getViewport();
      insertOffset.current += 1;
      return { x: -viewport.x + insertOffset.current * 20 + 400, y: -viewport.y + insertOffset.current * 20 };
    };

    const handleAddServiceNode = () => {
      const id = uuidv4();
      updateData(id, {
        type: "service",
        serviceName: "",
        image: "",
        ports: [],
        env: [],
        files: [],
        isValid: false,
      });
      addNodes({
        id,
        position: getNewNodePosition(),
        width: 650,
        style: { width: "650px" },
        type: "serviceNode",
        data: {},
      });
    };

    const handleAddArtifactNode = () => {
      const id = uuidv4();
      updateData(id, { type: "artifact", artifactName: "", files: {}, isValid: false });
      addNodes({
        id,
        position: getNewNodePosition(),
        width: 400,
        style: { width: "400px" },
        type: "artifactNode",
        data: {},
      });
    };

    const handleAddShellNode = () => {
      const id = uuidv4();
      updateData(id, {
        type: "shell",
        shellName: "",
        command: "",
        image: "",
        env: [],
        files: [],
        store: "",
        wait_enabled: "true",
        wait: "",
        isValid: false,
      });
      addNodes({
        id,
        position: getNewNodePosition(),
        width: 650,
        style: { width: "650px" },
        type: "shellNode",
        data: {},
      });
    };

    const handleAddExecNode = () => {
      const id = uuidv4();
      updateData(id, {
        type: "exec",
        execName: "",
        serviceName: "",
        command: "",
        acceptableCodes: [],
        isValid: false,
      });
      addNodes({
        id,
        position: getNewNodePosition(),
        width: 400,
        style: { width: "400px" },
        type: "execNode",
        data: {},
      });
    };

    const handleNodeDoubleClick = useCallback(
      (e: React.MouseEvent, node: Node) => {
        fitView({ nodes: [node], maxZoom: 1, duration: 500 });
      },
      [fitView],
    );

    useEffect(() => {
      setEdges((prevState) => {
        return Object.entries(getNodeDependencies(data)).flatMap(([to, froms]) =>
          [...froms].map((from) => ({
            id: `${from}-${to}`,
            source: from,
            target: to,
            animated: true,
            type: "straight",
            style: { strokeWidth: "3px" },
          })),
        );
      });
    }, [setEdges, data]);

    // Remove the resizeObserver error
    useEffect(() => {
      const errorHandler = (e: any) => {
        if (
          e.message.includes(
            "ResizeObserver loop completed with undelivered notifications" || "ResizeObserver loop limit exceeded",
          )
        ) {
          const resizeObserverErr = document.getElementById("webpack-dev-server-client-overlay");
          if (resizeObserverErr) {
            resizeObserverErr.style.display = "none";
          }
        }
      };
      window.addEventListener("error", errorHandler);

      return () => {
        window.removeEventListener("error", errorHandler);
      };
    }, []);

    useImperativeHandle(
      ref,
      () => ({
        getStarlark: () => {
          return generateStarlarkFromGraph(nodes, edges, data, existingEnclave);
        },
      }),
      [nodes, edges, data, existingEnclave],
    );

    return (
      <Flex flexDirection={"column"} h={"100%"} gap={"8px"}>
        <ButtonGroup paddingInline={6}>
          <Button onClick={onLayout}>Do Layout</Button>
          <Button leftIcon={<FiPlusCircle />} onClick={handleAddServiceNode}>
            Add Service Node
          </Button>
          <Button leftIcon={<FiPlusCircle />} onClick={handleAddArtifactNode}>
            Add Files Node
          </Button>
          <Button leftIcon={<FiPlusCircle />} onClick={handleAddShellNode}>
            Add Shell Node
          </Button>
          <Button leftIcon={<FiPlusCircle />} onClick={handleAddExecNode}>
            Add Exec Node
          </Button>
        </ButtonGroup>
        <Box bg={"gray.900"} flex={"1"}>
          <ReactFlow
            minZoom={0.1}
            maxZoom={1}
            nodes={nodes}
            edges={edges}
            proOptions={{ hideAttribution: true }}
            onMove={() => (insertOffset.current = 1)}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onNodeDoubleClick={handleNodeDoubleClick}
            nodeTypes={nodeTypes}
            fitView
          >
            <Controls />
            <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
          </ReactFlow>
        </Box>
      </Flex>
    );
  },
);
Visualiser.displayName = "ForwardRef Visualiser";
