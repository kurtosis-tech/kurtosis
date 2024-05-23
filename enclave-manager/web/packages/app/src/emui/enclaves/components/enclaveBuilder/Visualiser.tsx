import { Box, Flex } from "@chakra-ui/react";
import { RemoveFunctions } from "kurtosis-ui-components";
import { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  Edge,
  MarkerType,
  Node,
  ReactFlow,
  useEdgesState,
  useNodesState,
} from "reactflow";
import { EnclaveFullInfo } from "../../types";
import { KurtosisArtifactNode } from "./nodes/KurtosisArtifactNode";
import { KurtosisExecNode } from "./nodes/KurtosisExecNode";
import { KurtosisPackageNode } from "./nodes/KurtosisPackageNode";
import { KurtosisPythonNode } from "./nodes/KurtosisPythonNode";
import { KurtosisServiceNode } from "./nodes/KurtosisServiceNode";
import { KurtosisShellNode } from "./nodes/KurtosisShellNode";
import { Toolbar } from "./Toolbar";
import { useUIState } from "./UIStateContext";
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
    const { data, initialImportedPackageData } = useVariableContext();
    const insertOffset = useRef(0);
    const [nodes, , onNodesChange] = useNodesState(initialNodes || []);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges || []);
    // TODO(skylar): make this work for multiple packages?
    const [shouldApplyAutoLayout, setShouldApplyAutoLayout] = useState(false);
    const { expandedNodes, applyAutoLayout, zoomToNode } = useUIState();

    const handleNodeClick = useCallback(
      (e: React.MouseEvent, node: Node) => {
        // Only zoom to node if it is not expanded (i.e. it is about to be expanded)
        if (expandedNodes[node.id]) {
          return;
        }
        zoomToNode(node);
      },
      [zoomToNode, expandedNodes],
    );

    useEffect(() => {
      setEdges((prevState) => {
        const nextEdges = Object.entries(getNodeDependencies(data)).flatMap(([to, froms]) =>
          [...froms].map((from) => ({
            id: `${from}-${to}`,
            source: from,
            target: to,
          })),
        );
        if (nextEdges.length > 0) {
          setShouldApplyAutoLayout(true);
        }
        return nextEdges;
      });
    }, [setEdges, data, applyAutoLayout]);

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
          return generateStarlarkFromGraph(nodes, edges, data, initialImportedPackageData, existingEnclave);
        },
      }),
      [nodes, edges, data, initialImportedPackageData, existingEnclave],
    );

    useEffect(() => {
      if (shouldApplyAutoLayout) {
        applyAutoLayout();
      }
    }, [shouldApplyAutoLayout, applyAutoLayout]);

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
            // TODO(skylar): fix this. it currently zooms to a random place off the top of the graph
            onNodeClick={handleNodeClick}
            nodeTypes={nodeTypes}
            fitView
            snapToGrid
            onlyRenderVisibleElements={false} // This is required to prevent the package node from re-fetching data when it is re-rendered
            defaultEdgeOptions={{
              focusable: false,
              animated: false,
              style: { strokeWidth: "2px" },
              markerEnd: {
                type: MarkerType.ArrowClosed,
              },
            }}
          >
            <Controls />
            <Background variant={BackgroundVariant.Dots} gap={24} size={2} />
          </ReactFlow>
        </Box>
      </Flex>
    );
  },
);
Visualiser.displayName = "ForwardRef Visualiser";
