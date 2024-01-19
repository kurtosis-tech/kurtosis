import {
  Box,
  Button,
  Flex,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Spinner,
  Text,
} from "@chakra-ui/react";
import Dagre from "@dagrejs/dagre";
import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { isDefined, KurtosisAlert, RemoveFunctions } from "kurtosis-ui-components";
import { useCallback, useEffect, useState } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  Edge,
  Node,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
} from "reactflow";
import "reactflow/dist/style.css";
import { useEnclavesContext } from "../../EnclavesContext";

type DryRunVisualiseModalMode =
  | { type: "loading" }
  | { type: "ready"; runLines: StarlarkRunResponseLine[] }
  | { type: "error"; error: string }
  | { type: "closing" };

type DryRunVisualiseModalProps = {
  isOpen: boolean;
  onClose: () => void;
  enclave?: RemoveFunctions<EnclaveInfo>;
  packageId: string;
  args?: Record<string, any>;
};

export const DryRunVisualiseModal = ({
  isOpen,
  onClose,
  enclave: enclaveProp,
  args,
  packageId,
}: DryRunVisualiseModalProps) => {
  const { runStarlarkPackage, createEnclave, destroyEnclaves } = useEnclavesContext();
  const [mode, setMode] = useState<DryRunVisualiseModalMode>({ type: "loading" });
  const [enclave, setEnclave] = useState(enclaveProp);

  const handleClose = async () => {
    setMode({ type: "closing" });
    if (!isDefined(enclaveProp) && mode.type === "ready" && isDefined(enclave)) {
      await destroyEnclaves([enclave.enclaveUuid]);
    }
    onClose();
  };

  useEffect(() => {
    if (isOpen && isDefined(args)) {
      (async () => {
        let enclaveToDryRun = enclaveProp;
        if (!isDefined(enclaveToDryRun)) {
          const createEnclaveResponse = await createEnclave("", "info");
          if (createEnclaveResponse.isErr) {
            setMode({ type: "error", error: createEnclaveResponse.error });
            return;
          }
          if (!isDefined(createEnclaveResponse.value.enclaveInfo)) {
            setMode({ type: "error", error: "Create enclave succeeded, but no enclave was returned" });
            return;
          }
          enclaveToDryRun = createEnclaveResponse.value.enclaveInfo;
        }
        setEnclave(enclaveToDryRun);

        const starlarkResponse = await runStarlarkPackage(enclaveToDryRun, packageId, args, true);
        const runLines: StarlarkRunResponseLine[] = [];
        for await (const line of starlarkResponse) {
          runLines.push(line);
        }
        setMode({ type: "ready", runLines });
        console.log(runLines);
      })();
    }
  }, [isOpen, enclaveProp, args, packageId]);

  return (
    <Modal isOpen={isOpen} onClose={handleClose}>
      <ModalOverlay />
      <ModalContent h={"90vh"} minW={"1024px"}>
        <ModalHeader>Visualise Starlark Run</ModalHeader>
        <ModalCloseButton isDisabled={mode.type === "loading" || mode.type === "closing"} />
        <ModalBody>
          {mode.type === "loading" && <Spinner />}
          {mode.type === "ready" && (
            <ReactFlowProvider>
              <Visualiser runLines={mode.runLines} />
            </ReactFlowProvider>
          )}
          {mode.type === "error" && <KurtosisAlert message={mode.error} />}
          {mode.type === "closing" && (
            <Flex>
              <Spinner />
              <Text>Tidying up</Text>
            </Flex>
          )}
        </ModalBody>
        <ModalFooter>
          <Button onClick={handleClose} isDisabled={mode.type === "loading"} isLoading={mode.type === "closing"}>
            Close
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));

const getLayoutedElements = (nodes: Node<{ label: string }, string | undefined>[], edges: Edge<any>[]) => {
  g.setGraph({ rankdir: "TB" });

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

type VisualiserProps = {
  runLines: StarlarkRunResponseLine[];
};

const initialNodes = [
  { id: "1", position: { x: 0, y: 0 }, data: { label: "1" } },
  { id: "2", position: { x: 0, y: 100 }, data: { label: "2" } },
  { id: "3", position: { x: 50, y: 100 }, data: { label: "2" } },
];
const initialEdges = [
  { id: "e1-2", source: "1", target: "2" },
  { id: "e2-3", source: "2", target: "3" },
];
const Visualiser = ({ runLines }: VisualiserProps) => {
  const { fitView } = useReactFlow();
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  const onLayout = useCallback(() => {
    const layouted = getLayoutedElements(nodes, edges);

    setNodes([...layouted.nodes]);
    setEdges([...layouted.edges]);

    window.requestAnimationFrame(() => {
      fitView();
    });
  }, [nodes, edges]);

  return (
    <Box h={"100%"}>
      <Button onClick={onLayout}>Do Layout</Button>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onInit={onLayout}
        proOptions={{ hideAttribution: true }}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        fitView
      >
        <Controls />
        <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
      </ReactFlow>
    </Box>
  );
};
