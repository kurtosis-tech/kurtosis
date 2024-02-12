import { isDefined, RemoveFunctions, stringifyError } from "kurtosis-ui-components";
import { Edge, Node } from "reactflow";
import { Result } from "true-myth";
import { EnclaveFullInfo } from "../../../types";
import { KurtosisNodeData, KurtosisServiceNodeData, Variable } from "./types";

export const EMUI_BUILD_STATE_KEY = "EMUI_BUILD_STATE";

export function starlarkScriptContainsEMUIBuildState(script: string) {
  return script.includes(EMUI_BUILD_STATE_KEY);
}

export function getInitialGraphStateFromEnclave<T extends object>(
  enclave?: RemoveFunctions<EnclaveFullInfo>,
): Result<{ nodes: Node<any>[]; edges: Edge<any>[]; data: Record<string, T> }, string> {
  if (!isDefined(enclave)) {
    return Result.ok({ nodes: [], edges: [], data: {} });
  }
  if (!isDefined(enclave.starlarkRun)) {
    return Result.err("Enclave has no previous starlark run.");
  }
  if (enclave.starlarkRun.isErr) {
    return Result.err(`Fetching previous starlark run resulted in an error: ${enclave.starlarkRun.error}`);
  }
  const b64State = enclave.starlarkRun.value.serializedScript
    .split("\n")
    .find((line) => line.includes(EMUI_BUILD_STATE_KEY));
  if (!isDefined(b64State)) {
    return Result.err("Enclave wasn't created with the EMUI enclave builder.");
  }
  try {
    return Result.ok(JSON.parse(atob(b64State.split("=")[1])));
  } catch (error: any) {
    return Result.err(`Couldn't parse previous state: ${stringifyError(error)}`);
  }
}

export function getNodeName(kurtosisNodeData: KurtosisNodeData): string {
  if (kurtosisNodeData.type === "service") {
    return kurtosisNodeData.serviceName;
  }
  if (kurtosisNodeData.type === "artifact") {
    return kurtosisNodeData.artifactName;
  }
  if (kurtosisNodeData.type === "shell") {
    return kurtosisNodeData.shellName;
  }
  throw new Error(`Unknown node type.`);
}

function normaliseNameToStarlarkVariable(name: string) {
  return name.replace(/\s|-/g, "_").toLowerCase();
}

function escapeString(value: string): string {
  return value.replaceAll(/(["\\])/g, "\\$1");
}

const variablePattern = /\{\{((?:service|artifact|shell).([^.]+)\.?.*)}}/;
export function getVariablesFromNodes(nodes: Record<string, KurtosisNodeData>): Variable[] {
  return Object.entries(nodes).flatMap(([id, data]) => {
    if (data.type === "service") {
      return [
        {
          id: `service.${id}.name`,
          displayName: `service.${data.serviceName}.name`,
          value: `${normaliseNameToStarlarkVariable(data.serviceName)}.name`,
        },
        {
          id: `service.${id}.hostname`,
          displayName: `service.${data.serviceName}.hostname`,
          value: `${normaliseNameToStarlarkVariable(data.serviceName)}.hostname`,
        },
        ...data.ports.flatMap((port, i) => [
          {
            id: `service.${id}.port.${i}`,
            displayName: `service.${data.serviceName}.port.${port.portName}`,
            value: `"{}://{}:{}".format(${normaliseNameToStarlarkVariable(data.serviceName)}.ports["${
              port.portName
            }"].application_protocol, ${normaliseNameToStarlarkVariable(
              data.serviceName,
            )}.hostname, ${normaliseNameToStarlarkVariable(data.serviceName)}.ports["${port.portName}"].number)`,
          },
          {
            id: `service.${id}.port.${i}.port`,
            displayName: `service.${data.serviceName}.port.${port.portName}.port`,
            value: `${normaliseNameToStarlarkVariable(data.serviceName)}.ports["${port.portName}"].number`,
          },
          {
            id: `service.${id}.port.${i}.applicationProtocol`,
            displayName: `service.${data.serviceName}.port.${port.portName}.application_protocol`,
            value: `${normaliseNameToStarlarkVariable(data.serviceName)}.ports["${
              port.portName
            }"].application_protocol`,
          },
        ]),
        ...data.env.map((env, i) => ({
          id: `service.${id}.env.${i}`,
          displayName: `service.${data.serviceName}.env.${env.key}`,
          value: `"${env.value}"`,
        })),
      ];
    }
    if (data.type === "artifact") {
      return [
        {
          id: `artifact.${id}`,
          displayName: `artifact.${data.artifactName}`,
          value: `${normaliseNameToStarlarkVariable(data.artifactName)}`,
        },
      ];
    }

    if (data.type === "shell") {
      return [
        {
          id: `shell.${id}`,
          displayName: `shell.${data.shellName}`,
          value: `${normaliseNameToStarlarkVariable(data.shellName)}`,
        },
        ...data.env.map((env, i) => ({
          id: `shell.${id}.env.${i}`,
          displayName: `shell.${data.shellName}.env.${env.key}`,
          value: `"${env.value}"`,
        })),
      ];
    }

    return [];
  });
}

export function getNodeDependencies(nodes: Record<string, KurtosisNodeData>): Record<string, Set<string>> {
  const dependencies: Record<string, Set<string>> = {};
  const getDependenciesFor = (key: string): Set<string> => {
    if (!isDefined(dependencies[key])) {
      dependencies[key] = new Set<string>();
    }
    return dependencies[key];
  };
  Object.entries(nodes).forEach(([id, data]) => {
    if (data.type === "service") {
      const nameMatches = data.serviceName.match(variablePattern);
      if (nameMatches) {
        getDependenciesFor(id).add(nameMatches[2]);
      }
      data.env.forEach((env) => {
        const envMatches = env.key.match(variablePattern) || env.value.match(variablePattern);
        if (envMatches) {
          getDependenciesFor(id).add(envMatches[2]);
        }
      });
      data.ports.forEach((port) => {
        const portMatches = port.portName.match(variablePattern) || port.applicationProtocol.match(variablePattern);
        if (portMatches) {
          getDependenciesFor(id).add(portMatches[2]);
        }
      });
      data.files.forEach((file) => {
        const fileMatches = file.mountPoint.match(variablePattern) || file.artifactName.match(variablePattern);
        if (fileMatches) {
          getDependenciesFor(id).add(fileMatches[2]);
        }
      });
    }
    if (data.type === "shell") {
      const nameMatches = data.shellName.match(variablePattern);
      if (nameMatches) {
        getDependenciesFor(id).add(nameMatches[2]);
      }
      data.env.forEach((env) => {
        const envMatches = env.key.match(variablePattern) || env.value.match(variablePattern);
        if (envMatches) {
          getDependenciesFor(id).add(envMatches[2]);
        }
      });
      data.files.forEach((file) => {
        const fileMatches = file.mountPoint.match(variablePattern) || file.artifactName.match(variablePattern);
        if (fileMatches) {
          getDependenciesFor(id).add(fileMatches[2]);
        }
      });
    }
  });
  return dependencies;
}

export function generateStarlarkFromGraph(
  nodes: Node[],
  edges: Edge[],
  data: Record<string, KurtosisNodeData>,
  existingEnclave?: RemoveFunctions<EnclaveFullInfo>,
): string {
  // Topological sort
  const sortedNodes: Node<KurtosisServiceNodeData>[] = [];
  let remainingEdges = [...edges];
  while (remainingEdges.length > 0 || sortedNodes.length !== nodes.length) {
    const nodesRemoved = nodes
      .filter((node) => remainingEdges.every((edge) => edge.target !== node.id)) // eslint-disable-line no-loop-func
      .filter((node) => !sortedNodes.includes(node));
    if (nodesRemoved.length === 0) {
      throw new Error(
        "Topological sort couldn't remove nodes. This indicates a cycle has been detected. Starlark cannot be rendered.",
      );
    }
    sortedNodes.push(...nodesRemoved);
    remainingEdges = remainingEdges.filter((edge) => sortedNodes.every((node) => edge.source !== node.id));
  }
  const variablesById = getVariablesFromNodes(data).reduce(
    (acc, cur) => ({ ...acc, [cur.id]: cur }),
    {} as Record<string, Variable>,
  );
  const interpolateValue = (input: string): string => {
    let formatString = input;
    let variableMatches = formatString.match(variablePattern);
    if (!isDefined(variableMatches)) {
      return `"${formatString}"`;
    }

    const references: string[] = [];
    while (isDefined(variableMatches)) {
      formatString = formatString.replace(variableMatches[0], "{}");
      references.push(variablesById[variableMatches[1]].value);
      variableMatches = formatString.match(variablePattern);
    }
    if (formatString === "{}") {
      return references[0];
    }

    return `"${formatString}".format(${references.join(", ")})`;
  };

  let starlark = "def run(plan):\n";
  for (const node of sortedNodes) {
    const nodeData = data[node.id];
    if (nodeData.type === "service") {
      const serviceName = normaliseNameToStarlarkVariable(nodeData.serviceName);
      starlark += `    ${serviceName} = plan.add_service(\n`;
      starlark += `        name = ${interpolateValue(nodeData.serviceName)},\n`;
      starlark += `        config = ServiceConfig (\n`;
      starlark += `            image = ${interpolateValue(nodeData.image)},\n`;
      starlark += `            ports = {\n`;
      for (const { portName, port, applicationProtocol, transportProtocol } of nodeData.ports) {
        starlark += `                ${interpolateValue(portName)}: PortSpec(\n`;
        starlark += `                    number = ${port},\n`;
        starlark += `                    transport_protocol = "${transportProtocol}",\n`;
        starlark += `                    application_protocol = ${interpolateValue(applicationProtocol)},\n`;
        starlark += `                ),\n`;
      }
      starlark += `            },\n`;
      starlark += `            env_vars = {\n`;
      for (const { key, value } of nodeData.env) {
        starlark += `                ${interpolateValue(key)}: ${interpolateValue(value)},\n`;
      }
      starlark += `            },\n`;
      starlark += `            files = {\n`;
      for (const { mountPoint, artifactName } of nodeData.files) {
        starlark += `                ${interpolateValue(mountPoint)}: ${interpolateValue(artifactName)},\n`;
      }
      starlark += `            },\n`;
      starlark += `        ),\n`;
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "artifact") {
      const artifactName = normaliseNameToStarlarkVariable(nodeData.artifactName);
      starlark += `    ${artifactName} = plan.render_templates(\n`;
      starlark += `        name = "${nodeData.artifactName}",\n`;
      starlark += `        config = {\n`;
      for (const [fileName, fileText] of Object.entries(nodeData.files)) {
        starlark += `            "${fileName}": struct(\n`;
        starlark += `                template="""${escapeString(fileText)}""",\n`;
        starlark += `                data={},\n`;
        starlark += `            ),\n`;
      }
      starlark += `        },\n`;
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "shell") {
      const shellName = normaliseNameToStarlarkVariable(nodeData.shellName);
      starlark += `    ${shellName} = plan.run_sh(\n`;
      starlark += `        run = """${escapeString(nodeData.command)}""",\n`;
      const image = interpolateValue(nodeData.image);
      if (image !== '""') {
        starlark += `        image = ${image},\n`;
      }
      starlark += `        env_vars = {\n`;
      for (const { key, value } of nodeData.env) {
        starlark += `            ${interpolateValue(key)}: ${interpolateValue(value)},\n`;
      }
      starlark += `        },\n`;
      starlark += `        files = {\n`;
      for (const { mountPoint, artifactName } of nodeData.files) {
        starlark += `            ${interpolateValue(mountPoint)}: ${interpolateValue(artifactName)},\n`;
      }
      starlark += `        },\n`;
      starlark += `        store = [\n`;
      for (const store of nodeData.store) {
        starlark += `            ${interpolateValue(store.value)},\n`;
      }
      starlark += `        ],\n`;
      const wait = interpolateValue(nodeData.wait);
      if (nodeData.wait_enabled === "false" || wait !== '""') {
        starlark += `        wait=${nodeData.wait_enabled === "true" ? wait : "None"},\n`;
      }
      starlark += `    )\n\n`;
    }
  }

  // Delete any services from any existing enclave that aren't defined anymore
  if (isDefined(existingEnclave) && existingEnclave.services?.isOk) {
    for (const existingService of Object.values(existingEnclave.services.value.serviceInfo)) {
      const serviceNoLongerExists = sortedNodes.every((node) => {
        const nodeData = data[node.id];
        return nodeData.type !== "service" || nodeData.serviceName !== existingService.name;
      });
      if (serviceNoLongerExists) {
        starlark += `    plan.remove_service(name = "${existingService.name}")\n`;
      }
    }
  }

  starlark += `\n\n# ${EMUI_BUILD_STATE_KEY}=${btoa(JSON.stringify({ nodes, edges, data }))}`;

  console.log(starlark);

  return starlark;
}
