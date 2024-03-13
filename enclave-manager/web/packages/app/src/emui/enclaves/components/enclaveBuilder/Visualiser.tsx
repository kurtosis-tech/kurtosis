import { Box, Flex } from "@chakra-ui/react";
import { RemoveFunctions } from "kurtosis-ui-components";
import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef } from "react";
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
} from "reactflow";
import { EnclaveFullInfo } from "../../types";
import { KurtosisArtifactNode } from "./nodes/KurtosisArtifactNode";
import { KurtosisExecNode } from "./nodes/KurtosisExecNode";
import { KurtosisPackageNode } from "./nodes/KurtosisPackageNode";
import { KurtosisPythonNode } from "./nodes/KurtosisPythonNode";
import { KurtosisServiceNode } from "./nodes/KurtosisServiceNode";
import { KurtosisShellNode } from "./nodes/KurtosisShellNode";
import { Toolbar } from "./Toolbar";
import { generateStarlarkFromGraph, getNodeDependencies } from "./utils";
import { useVariableContext } from "./VariableContextProvider";
import "./Visualiser.css";

const nodeTypes = {
  serviceNode: KurtosisServiceNode,
  execNode: KurtosisExecNode,
  artifactNode: KurtosisArtifactNode,
  shellNode: KurtosisShellNode,
  pythonNode: KurtosisPythonNode,
  packageNode: KurtosisPackageNode,
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
    const { data } = useVariableContext();
    const insertOffset = useRef(0);
    const { fitView } = useReactFlow();
    const [nodes, , onNodesChange] = useNodesState(initialNodes || []);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges || []);

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
      <Flex position="relative" flexDirection={"column"} h={"100%"} gap={"8px"}>
        <Toolbar />
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
