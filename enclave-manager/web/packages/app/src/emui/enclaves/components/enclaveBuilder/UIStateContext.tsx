import { Edge, Node, useReactFlow } from "reactflow";

import Dagre from "@dagrejs/dagre";
import { createContext, PropsWithChildren, useCallback, useContext, useState } from "react";

// Define the context
const UIContext = createContext<{
  expandedNodes: Record<string, boolean>;
  toggleExpanded: (nodeId: string) => void;
  applyAutoLayout: () => void;
}>({
  expandedNodes: {},
  toggleExpanded: () => {},
  applyAutoLayout: () => {},
});

// Graph layout for auto-layouting the nodes
const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));
const getLayoutedElements = <T extends object>(nodes: Node<T>[], edges: Edge<any>[]) => {
  if (nodes.length === 0) {
    return { nodes, edges };
  }
  g.setGraph({ rankdir: "LR", ranksep: 200, nodesep: 200 });

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

// Define the provider
export const UIStateProvider = ({ children }: PropsWithChildren) => {
  const [expandedNodes, setExpandedNodes] = useState<Record<string, boolean>>({});

  const { fitView, getNodes, getEdges, setNodes, setEdges } = useReactFlow();

  // Toggles whether a node is expanded or not. A node only appears expanded if
  // the zoom level is zoomed in enough (see ZoomAwareNodeContent)
  const toggleExpanded = useCallback(
    (nodeId: string) => {
      setExpandedNodes({ ...expandedNodes, [nodeId]: !expandedNodes[nodeId] });
    },
    [expandedNodes, setExpandedNodes],
  );

  // Re-layouts the graph
  const applyAutoLayout = useCallback(() => {
    const nodes = getNodes();
    const edges = getEdges();
    const layouted = getLayoutedElements(nodes, edges);

    setNodes([...layouted.nodes]);
    setEdges([...layouted.edges]);

    window.requestAnimationFrame(() => {
      fitView();
    });
  }, [fitView, setEdges, setNodes, getEdges, getNodes]);

  return <UIContext.Provider value={{ expandedNodes, toggleExpanded, applyAutoLayout }}>{children}</UIContext.Provider>;
};

// Define the hook
export function useUIState() {
  return useContext(UIContext);
}
