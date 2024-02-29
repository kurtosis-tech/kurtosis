import { Button, Flex } from "@chakra-ui/react";
import { isDefined } from "kurtosis-ui-components";
import { memo, useEffect, useState } from "react";
import { FiEdit } from "react-icons/fi";
import { NodeProps, useReactFlow } from "reactflow";
import YAML from "yaml";
import { useKurtosisClient } from "../../../../../client/enclaveManager/KurtosisClientContext";
import { KurtosisFormControl } from "../../form/KurtosisFormControl";
import { StringArgumentInput } from "../../form/StringArgumentInput";
import { validateName } from "../input/validators";
import { ConfigurePackageNodeModal } from "../modals/ConfigurePackageNodeModal";
import { KurtosisPackageNodeData, PlanYaml } from "../types";
import { useVariableContext } from "../VariableContextProvider";
import { KurtosisNode } from "./KurtosisNode";

type Mode = { type: "loading" } | { type: "error"; error: string } | { type: "ready" };

export const KurtosisPackageNode = memo(
  ({ id, selected }: NodeProps) => {
    const { getNodes, deleteElements, setNodes } = useReactFlow();
    const [showPackageConfigModal, setShowPackageConfigModal] = useState(false);
    const [mode, setMode] = useState<Mode>({ type: "ready" });
    const kurtosisClient = useKurtosisClient();
    const { data, updateData, removeData } = useVariableContext();
    const nodeData = data[id] as KurtosisPackageNodeData | undefined;

    useEffect(() => {
      const packageId = nodeData?.packageId;
      const args = nodeData?.args;
      if (isDefined(packageId) && isDefined(args) && packageId !== "") {
        let cancelled = false;
        (async () => {
          setMode({ type: "loading" });
          const enclave = await kurtosisClient.createEnclave("", "info");
          if (enclave.isErr) {
            setMode({ type: "error", error: enclave.error });
            return;
          }
          if (!isDefined(enclave.value.enclaveInfo) || !isDefined(enclave.value.enclaveInfo.apiContainerInfo)) {
            setMode({ type: "error", error: "APIC info missing from temporary enclave" });
            return;
          }
          const plan = await kurtosisClient.getStarlarkPackagePlanYaml(
            enclave.value.enclaveInfo.apiContainerInfo,
            packageId,
            args,
          );
          await kurtosisClient.destroy(enclave.value.enclaveInfo?.enclaveUuid);
          if (cancelled) {
            return;
          }
          if (plan.isErr) {
            setMode({ type: "error", error: plan.error });
            return;
          }
          console.log(plan.value.planYaml);
          const parsedPlan = YAML.parse(plan.value.planYaml) as PlanYaml;

          // Remove current children
          const nodesToRemove = getNodes().filter((node) => node.parentNode === id);
          deleteElements({ nodes: nodesToRemove });
          removeData(nodesToRemove);
          const nodesToAdd: { type: string; id: string }[] = [
            ...parsedPlan.services.map((service, i) => ({
              type: "serviceNode",
              id: `${id}:${service.uuid}`,
            })),
            ...parsedPlan.tasks.map((task, i) => ({
              type: task.taskType === "exec" ? "execNode" : task.taskType === "python" ? "pythonNode" : "shellNode",
              id: `${id}:${task.uuid}`,
            })),
          ];
          const serviceNamesToId = parsedPlan.services.reduce(
            (acc: Record<string, string>, service) => ({ ...acc, [service.name]: `${id}:${service.uuid}` }),
            {},
          );
          parsedPlan.services.forEach((service) =>
            updateData(`${id}:${service.uuid}`, {
              type: "service",
              name: service.name,
              isFromPackage: true,
              env: service.envVars || [],
              image: {
                type: "image",
                image: service.image.name,
                registryUsername: "",
                registryPassword: "",
                registry: "",
                buildContextDir: "",
                targetStage: "",
                flakeLocationDir: "",
                flakeOutput: "",
              },
              ports: (service.ports || []).map((port) => ({
                name: port.name,
                port: port.number,
                applicationProtocol: port.applicationProtocol || "",
                transportProtocol: port.transportProtocol,
              })),
              isValid: true,
              files: [],
              cmd: (service.command || []).join(" "),
              entrypoint: (service.entrypoint || []).join(" "),
            }),
          );
          parsedPlan.tasks.forEach((task) => {
            if (task.taskType === "exec") {
              const serviceVariable = `{{service.${serviceNamesToId[task.serviceName]}.name}}`;
              updateData(`${id}:${task.uuid}`, {
                type: "exec",
                name: "",
                isValid: true,
                isFromPackage: true,
                service: serviceVariable,
                command: task.command,
                acceptableCodes: (task.acceptableCodes || []).map((code) => ({ value: code })),
              });
            }
          });
          setNodes((nodes) => [
            ...nodes,
            ...nodesToAdd.map((node, i) => ({
              ...node,
              parentNode: id,
              data: {},
              extent: "parent" as "parent",
              position: { x: 50 + 700 * (i % 3), y: 200 + 700 * Math.floor(i / 3) },
            })),
          ]);

          setMode({ type: "ready" });
        })();
        return () => {
          cancelled = true;
        };
      }
    }, [
      nodeData?.packageId,
      nodeData?.args,
      deleteElements,
      getNodes,
      id,
      kurtosisClient,
      removeData,
      setNodes,
      updateData,
    ]);

    if (!isDefined(nodeData)) {
      return null;
    }

    return (
      <KurtosisNode
        id={id}
        selected={selected}
        minWidth={900}
        maxWidth={10000}
        portalContent={
          <ConfigurePackageNodeModal
            isOpen={showPackageConfigModal}
            onClose={() => setShowPackageConfigModal(false)}
            initialValues={nodeData.args}
          />
        }
        backgroundColor={"transparent"}
      >
        <Flex gap={"16px"}>
          <KurtosisFormControl<KurtosisPackageNodeData> name={"name"} label={"Node Name"} isRequired flex={"1"}>
            <StringArgumentInput size={"sm"} name={"name"} isRequired validate={validateName} />
          </KurtosisFormControl>
          <KurtosisFormControl<KurtosisPackageNodeData>
            name={"packageId"}
            label={`Package ${nodeData.packageId}`}
            isRequired
            flex={"1"}
          >
            <Button
              w={"100%"}
              size={"sm"}
              leftIcon={<FiEdit />}
              onClick={() => setShowPackageConfigModal(true)}
              isLoading={mode.type === "loading"}
            >
              Edit
            </Button>
          </KurtosisFormControl>
        </Flex>
      </KurtosisNode>
    );
  },
  (oldProps, newProps) => oldProps.id === newProps.id && oldProps.selected === newProps.selected,
);
