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
import { KurtosisPackageNodeData, PlanFileArtifact, PlanTask, PlanYaml } from "../types";
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
      let args = nodeData?.args;
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

          // If args only has one record, and its value is args, ASSUME user is passing args via a JSON or YAML into the args object of def run(plan, args)
          // via Json or Yaml editor
          // If only an `args` object is provided, kurtosis will not interpret the value in the args object as passing args via the args dictionary is (technically) deprecated even though it's still allowed
          if (Object.keys(args).length === 1 && args.hasOwnProperty("args")) {
            args = args["args"] as Record<string, string>; // TODO(tedi): ideally we'd validate and handle this in transform args utils
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

          const serviceNamesToId = (parsedPlan.services || []).reduce(
            (acc: Record<string, string>, service) => ({ ...acc, [service.name]: `${id}:${service.uuid}` }),
            {},
          );
          const taskLookup = (parsedPlan.tasks || []).reduce(
            (acc: Record<string, PlanTask>, task) => ({ ...acc, [task.uuid]: task }),
            {},
          );
          const artifactLookup = (parsedPlan.filesArtifacts || []).reduce(
            (acc: Record<string, PlanFileArtifact>, filesArtifact) => ({ ...acc, [filesArtifact.uuid]: filesArtifact }),
            {},
          );

          const plannedArtifacts = (parsedPlan.filesArtifacts || []).filter(
            (artifact) =>
              !(parsedPlan.tasks || []).some(
                (task) => task.taskType !== "exec" && (task.store || []).some((store) => store.uuid === artifact.uuid),
              ),
          );

          const artifactToNodeId = (parsedPlan.filesArtifacts || []).reduce(
            (acc: Record<string, string>, artifact) => ({
              ...acc,
              [artifact.uuid]:
                parsedPlan.tasks?.find(
                  (task) => task.taskType !== "exec" && task.store?.some((store) => store.uuid === artifact.uuid),
                )?.uuid || artifact.uuid,
            }),
            {},
          );

          const artifactTypes = (parsedPlan.filesArtifacts || []).reduce(
            (acc: Record<string, string>, artifact) => ({
              ...acc,
              [artifact.uuid]:
                taskLookup[artifactToNodeId[artifact.uuid]]?.taskType === "sh"
                  ? "shell"
                  : taskLookup[artifactToNodeId[artifact.uuid]]?.taskType === "python"
                  ? "python"
                  : "artifact",
            }),
            {},
          );

          const nodesToAdd: { type: string; id: string }[] = [
            ...(parsedPlan.services || []).map((service, i) => ({
              type: "serviceNode",
              id: `${id}:${service.uuid}`,
            })),
            ...(parsedPlan.tasks || []).map((task, i) => ({
              type: task.taskType === "exec" ? "execNode" : task.taskType === "python" ? "pythonNode" : "shellNode",
              id: `${id}:${task.uuid}`,
            })),
            ...plannedArtifacts.map((artifact, i) => ({
              type: "artifactNode",
              id: `${id}:${artifact.uuid}`,
            })),
          ];

          const futureReferencePattern = /\{\{\s*kurtosis\.([^.]+)\.(\S+?)\s*}}/;
          const convertFutureReferences = (input: string): string => {
            // All future references are assumed to be for services
            let result = input;
            let match = result.match(futureReferencePattern);
            while (isDefined(match)) {
              result = result.replaceAll(match[0], `{{service.${id}:${match[1]}.${match[2]}}}`);
              match = result.match(futureReferencePattern);
            }
            return result;
          };

          (parsedPlan.services || []).forEach((service) =>
            updateData(`${id}:${service.uuid}`, {
              type: "service",
              name: service.name,
              isFromPackage: true,
              env: (service.envVars || []).map(({ key, value }) => ({
                key: convertFutureReferences(key),
                value: convertFutureReferences(value),
              })),
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
              files: (service.files || []).flatMap((file) =>
                file.filesArtifacts.map((artifact) => ({
                  name: `{{${artifactTypes[artifact.uuid]}.${id}:${artifactToNodeId[artifact.uuid]}.store.${
                    artifact.name
                  }}}`,
                  mountPoint: file.mountPath,
                })),
              ),
              cmd: convertFutureReferences((service.command || []).join(" ")),
              entrypoint: convertFutureReferences((service.entrypoint || []).join(" ")),
            }),
          );
          (parsedPlan.tasks || []).forEach((task) => {
            if (task.taskType === "exec") {
              const serviceVariable = `{{service.${serviceNamesToId[task.serviceName]}.name}}`;
              updateData(`${id}:${task.uuid}`, {
                type: "exec",
                name: "",
                isValid: true,
                isFromPackage: true,
                service: serviceVariable,
                command: (task.command || []).join(" "),
                acceptableCodes: (task.acceptableCodes || []).map((code) => ({ value: code })),
              });
            }
            if (task.taskType === "python") {
              updateData(`${id}:${task.uuid}`, {
                type: "python",
                name: `Python ${task.uuid}`,
                isValid: true,
                isFromPackage: true,
                command: (task.command || []).join(" "),
                image: {
                  type: "image",
                  image: task.image,
                  registryUsername: "",
                  registryPassword: "",
                  registry: "",
                  buildContextDir: "",
                  targetStage: "",
                  flakeLocationDir: "",
                  flakeOutput: "",
                },
                packages: [],
                args: task.pythonArgs.map((arg) => ({ arg })),
                files: (task.files || []).flatMap((file) =>
                  file.filesArtifacts.map((artifact) => ({
                    name: `{{${artifactTypes[artifact.uuid]}.${id}:${artifactToNodeId[artifact.uuid]}.store.${
                      artifact.name
                    }}}`,
                    mountPoint: file.mountPath,
                  })),
                ),
                store: (task.store || []).map((store) => ({
                  name: store.name,
                  path: artifactLookup[store.uuid].files[0],
                })),
                wait_enabled: "false",
                wait: "",
              });
            }
            if (task.taskType === "sh") {
              updateData(`${id}:${task.uuid}`, {
                type: "shell",
                name: `Shell ${task.uuid}`,
                isValid: true,
                isFromPackage: true,
                command: (task.command || []).join(" "),
                image: {
                  type: "image",
                  image: task.image,
                  registryUsername: "",
                  registryPassword: "",
                  registry: "",
                  buildContextDir: "",
                  targetStage: "",
                  flakeLocationDir: "",
                  flakeOutput: "",
                },
                env: [],
                files: (task.files || []).flatMap((file) =>
                  file.filesArtifacts.map((artifact) => ({
                    name: `{{${artifactTypes[artifact.uuid]}.${id}:${artifactToNodeId[artifact.uuid]}.store.${
                      artifact.name
                    }}}`,
                    mountPoint: file.mountPath,
                  })),
                ),
                store: (task.store || []).map((store) => ({
                  name: store.name,
                  path: artifactLookup[store.uuid].files[0],
                })),
                wait_enabled: "false",
                wait: "",
              });
            }
          });
          plannedArtifacts.forEach((artifact) =>
            updateData(`${id}:${artifact.uuid}`, {
              type: "artifact",
              name: artifact.name,
              isFromPackage: true,
              isValid: true,
              files: artifact.files.reduce((acc, file) => ({ ...acc, [file]: "" }), {}),
            }),
          );

          setNodes((nodes) => [
            ...nodes,
            ...nodesToAdd.map((node, i) => ({
              ...node,
              parentNode: id,
              data: {},
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
