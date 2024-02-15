import {
  Button,
  ButtonGroup,
  Flex,
  ListItem,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Text,
  Tooltip,
  UnorderedList,
} from "@chakra-ui/react";
import { isDefined, KurtosisAlert, KurtosisAlertModal, RemoveFunctions, stringifyError } from "kurtosis-ui-components";
import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Edge, Node, ReactFlowProvider } from "reactflow";
import "reactflow/dist/style.css";
import { useEnclavesContext } from "../../EnclavesContext";
import { EnclaveFullInfo } from "../../types";
import { ViewStarlarkModal } from "./enclaveBuilder/modals/ViewStarlarkModal";
import { KurtosisNodeData } from "./enclaveBuilder/types";
import { getInitialGraphStateFromEnclave, getNodeName } from "./enclaveBuilder/utils";
import { useVariableContext, VariableContextProvider } from "./enclaveBuilder/VariableContextProvider";
import { Visualiser, VisualiserImperativeAttributes } from "./enclaveBuilder/Visualiser";

type EnclaveBuilderModalProps = {
  isOpen: boolean;
  onClose: () => void;
  existingEnclave?: RemoveFunctions<EnclaveFullInfo>;
};

export const EnclaveBuilderModal = (props: EnclaveBuilderModalProps) => {
  const variableContextKey = useRef(0);
  const [error, setError] = useState<string>();

  const {
    nodes: initialNodes,
    edges: initialEdges,
    data: initialData,
  } = useMemo((): {
    nodes: Node<any>[];
    edges: Edge<any>[];
    data: Record<string, KurtosisNodeData>;
  } => {
    variableContextKey.current += 1;
    const parseResult = getInitialGraphStateFromEnclave<KurtosisNodeData>(props.existingEnclave);
    if (parseResult.isErr) {
      setError(parseResult.error);
      return { nodes: [], edges: [], data: {} };
    }
    return {
      ...parseResult.value,
      data: Object.entries(parseResult.value.data)
        .filter(([id, data]) => parseResult.value.nodes.some((node) => node.id === id))
        .reduce((acc, [id, data]) => ({ ...acc, [id]: data }), {} as Record<string, KurtosisNodeData>),
    };
  }, [props.existingEnclave]);

  useEffect(() => {
    if (!props.isOpen) {
      variableContextKey.current += 1;
    }
  }, [props.isOpen]);

  if (isDefined(error)) {
    return (
      <KurtosisAlertModal
        title={"Error"}
        content={error}
        isOpen={true}
        onClose={() => {
          setError(undefined);
          props.onClose();
        }}
      />
    );
  }

  return (
    <VariableContextProvider key={variableContextKey.current} initialData={initialData}>
      <EnclaveBuilderModalImpl {...props} initialNodes={initialNodes} initialEdges={initialEdges} />
    </VariableContextProvider>
  );
};

type EnclaveBuilderModalImplProps = EnclaveBuilderModalProps & {
  initialNodes: Node[];
  initialEdges: Edge[];
};
const EnclaveBuilderModalImpl = ({
  isOpen,
  onClose,
  existingEnclave,
  initialNodes,
  initialEdges,
}: EnclaveBuilderModalImplProps) => {
  const navigator = useNavigate();
  const visualiserRef = useRef<VisualiserImperativeAttributes | null>(null);
  const { createEnclave, runStarlarkScript } = useEnclavesContext();
  const { data } = useVariableContext();
  const dataIssues = useMemo(
    () =>
      Object.values(data)
        .filter((nodeData) => !nodeData.isValid)
        .map((nodeData) => `${nodeData.type} ${getNodeName(nodeData)} has invalid data`),
    [data],
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>();
  const [currentStarlarkPreview, setCurrentStarlarkPreview] = useState<string>();

  const handleRun = async () => {
    if (!isDefined(visualiserRef.current)) {
      setError("Cannot run when no services are defined");
      return;
    }

    setError(undefined);
    let enclave = existingEnclave;
    let enclaveUUID = existingEnclave?.shortenedUuid;
    if (!isDefined(existingEnclave)) {
      setIsLoading(true);
      const newEnclave = await createEnclave("", "info", true);
      setIsLoading(false);

      if (newEnclave.isErr) {
        setError(`Could not create enclave, got: ${newEnclave.error}`);
        return;
      }
      if (!isDefined(newEnclave.value.enclaveInfo)) {
        setError(`Did not receive enclave info when running createEnclave`);
        return;
      }
      enclave = newEnclave.value.enclaveInfo;
      enclaveUUID = newEnclave.value.enclaveInfo.shortenedUuid;
    }

    if (!isDefined(enclave)) {
      setError(`Cannot trigger starlark run as enclave info cannot be found`);
      return;
    }

    try {
      const logsIterator = await runStarlarkScript(enclave, visualiserRef.current.getStarlark(), {});
      onClose();
      navigator(`/enclave/${enclaveUUID}/logs`, { state: { logs: logsIterator } });
    } catch (error: any) {
      setError(stringifyError(error));
    }
  };

  const handlePreview = () => {
    setCurrentStarlarkPreview(visualiserRef.current?.getStarlark() || "Unable to render");
  };

  return (
    <Modal isOpen={isOpen} onClose={!isLoading ? onClose : () => null} closeOnEsc={false}>
      <ModalOverlay />
      <ModalContent h={"89vh"} minW={"1300px"}>
        <ModalHeader>
          {isDefined(existingEnclave) ? `Editing ${existingEnclave.name}` : "Build a new Enclave"}
        </ModalHeader>
        <ModalCloseButton />
        <ModalBody paddingInline={"0"}>
          {isDefined(error) && <KurtosisAlert message={error} />}
          <ReactFlowProvider>
            <Visualiser
              ref={visualiserRef}
              initialNodes={initialNodes}
              initialEdges={initialEdges}
              existingEnclave={existingEnclave}
            />
          </ReactFlowProvider>
        </ModalBody>
        <ModalFooter>
          <ButtonGroup>
            <Button onClick={onClose} isDisabled={isLoading}>
              Close
            </Button>
            <Button onClick={handlePreview}>Preview</Button>
            <Tooltip
              label={
                dataIssues.length === 0 ? undefined : (
                  <Flex flexDirection={"column"}>
                    <Text>There are data issues that must be addressed before this enclave can run:</Text>
                    <UnorderedList>
                      {dataIssues.map((issue, i) => (
                        <ListItem key={i}>{issue}</ListItem>
                      ))}
                    </UnorderedList>
                  </Flex>
                )
              }
            >
              <Button
                onClick={handleRun}
                colorScheme={"green"}
                isLoading={isLoading}
                loadingText={"Run"}
                isDisabled={dataIssues.length > 0}
              >
                Run
              </Button>
            </Tooltip>
          </ButtonGroup>
        </ModalFooter>
      </ModalContent>
      <ViewStarlarkModal
        isOpen={isDefined(currentStarlarkPreview)}
        onClose={() => setCurrentStarlarkPreview(undefined)}
        starlark={currentStarlarkPreview}
      />
    </Modal>
  );
};
