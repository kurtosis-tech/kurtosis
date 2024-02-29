import { Box, Button, ButtonGroup, Icon } from "@chakra-ui/react";
import Dagre from "@dagrejs/dagre";
import { useCallback, useRef } from "react";
import { FiShare2 } from "react-icons/fi";
import { Edge, Node, useOnViewportChange, useReactFlow, XYPosition } from "reactflow";
import { v4 as uuidv4 } from "uuid";
import { nodeIcons } from "./nodes/KurtosisNode";
import { useVariableContext } from "./VariableContextProvider";

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

export const Toolbar = () => {
  const insertOffset = useRef(0);
  const { updateData } = useVariableContext();
  const { fitView, getViewport, getNodes, getEdges, addNodes, setNodes, setEdges } = useReactFlow();

  useOnViewportChange({ onEnd: () => (insertOffset.current = 1) });

  const onLayout = useCallback(() => {
    const nodes = getNodes();
    const edges = getEdges();
    const layouted = getLayoutedElements(nodes, edges);

    setNodes([...layouted.nodes]);
    setEdges([...layouted.edges]);

    window.requestAnimationFrame(() => {
      fitView();
    });
  }, [fitView, setEdges, setNodes, getEdges, getNodes]);

  const getNewNodePosition = (): XYPosition => {
    const viewport = getViewport();
    insertOffset.current += 1;
    return { x: -viewport.x + insertOffset.current * 20 + 400, y: -viewport.y + insertOffset.current * 20 };
  };

  const handleAddServiceNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "service",
      name: "",
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
      ports: [],
      env: [],
      files: [],
      cmd: "",
      entrypoint: "",
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

  const handleAddExecNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "exec",
      name: "",
      service: "",
      command: "",
      acceptableCodes: [],
      isValid: false,
    });
    addNodes({
      id,
      position: getNewNodePosition(),
      width: 650,
      style: { width: "650px" },
      type: "execNode",
      data: {},
    });
  };

  const handleAddArtifactNode = () => {
    const id = uuidv4();
    updateData(id, { type: "artifact", name: "", files: {}, isValid: false });
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
      name: "",
      command: "",
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
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

  const handleAddPythonNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "python",
      name: "",
      command: "",
      packages: [],
      image: {
        image: "",
        type: "image",
        buildContextDir: "",
        flakeLocationDir: "",
        flakeOutput: "",
        registry: "",
        registryPassword: "",
        registryUsername: "",
        targetStage: "",
      },
      args: [],
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
      type: "pythonNode",
      data: {},
    });
  };

  const handleAddPackageNode = () => {
    const id = uuidv4();
    updateData(id, {
      type: "package",
      name: "",
      packageId: "",
      args: {},
      isValid: false,
    });
    addNodes({
      id,
      position: getNewNodePosition(),
      width: 900,
      style: { width: "900px" },
      type: "packageNode",
      data: {},
    });
  };

  return (
    <Box
      borderRadius={"5px"}
      position={"absolute"}
      zIndex={"99999"}
      top={"20px"}
      left={"20px"}
      bg={"gray.800"}
      p={"8px"}
    >
      <ButtonGroup size={"sm"}>
        <Button leftIcon={<Icon as={FiShare2} />} onClick={onLayout}>
          Auto-Layout
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["service"]} />} onClick={handleAddServiceNode}>
          Add Service Node
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["artifact"]} />} onClick={handleAddArtifactNode}>
          Add Files Node
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["exec"]} />} onClick={handleAddExecNode}>
          Add Exec Node
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["shell"]} />} onClick={handleAddShellNode}>
          Add Shell Node
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["python"]} />} onClick={handleAddPythonNode}>
          Add Python Node
        </Button>
        <Button leftIcon={<Icon as={nodeIcons["package"]} />} onClick={handleAddPackageNode}>
          Add Package Node
        </Button>
      </ButtonGroup>
    </Box>
  );
};
