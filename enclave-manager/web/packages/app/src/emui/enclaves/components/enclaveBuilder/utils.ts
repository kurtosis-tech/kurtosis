import { isDefined, RemoveFunctions, stringifyError } from "kurtosis-ui-components";
import { Edge, Node } from "reactflow";
import { Result } from "true-myth";
import { EnclaveFullInfo } from "../../types";
import {
  KurtosisImageConfig,
  KurtosisNodeData,
  KurtosisPackageNodeData,
  KurtosisServiceNodeData,
  Variable,
} from "./types";

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

function normaliseNameToStarlarkVariable(name: string) {
  return name.replace(/\s|-/g, "_").toLowerCase();
}

function escapeString(value: string): string {
  return value.replaceAll(/(["\\])/g, "\\$1");
}

function escapeTemplateString(value: string): string {
  return escapeString(value).replaceAll(/{{(.*?)}}/g, "{{`{{$1}}`}}");
}

const variablePattern = /\{\{((?:service|artifact|shell|python).([^.]+)\.?.*?)}}/;
export function getVariablesFromNodes(nodes: Record<string, KurtosisNodeData>): Variable[] {
  return Object.entries(nodes).flatMap(([id, data]) => {
    if (data.type === "service") {
      return [
        {
          id: `service.${id}.name`,
          displayName: `${data.name}`,
          value: `${normaliseNameToStarlarkVariable(data.name)}.name`,
        },
        {
          id: `service.${id}.hostname`,
          displayName: `${data.name}.hostname`,
          value: `${normaliseNameToStarlarkVariable(data.name)}.hostname`,
        },
        ...data.ports.flatMap((port, i) => [
          {
            id: `service.${id}.ports.${i}`,
            displayName: `${data.name}.ports.${port.name}`,
            value: `"{}://{}:{}".format(${normaliseNameToStarlarkVariable(data.name)}.ports["${
              port.name
            }"].application_protocol, ${normaliseNameToStarlarkVariable(
              data.name,
            )}.hostname, ${normaliseNameToStarlarkVariable(data.name)}.ports["${port.name}"].number)`,
          },
          {
            id: `service.${id}.ports.${i}.port`,
            displayName: `${data.name}.ports.${port.name}.port`,
            value: `str(${normaliseNameToStarlarkVariable(data.name)}.ports["${port.name}"].number)`,
          },
          {
            id: `service.${id}.ports.${i}.applicationProtocol`,
            displayName: `${data.name}.ports.${port.name}.application_protocol`,
            value: `${normaliseNameToStarlarkVariable(data.name)}.ports["${port.name}"].application_protocol`,
          },
        ]),
        ...data.env.map((env, i) => ({
          id: `service.${id}.env.${i}.value`,
          displayName: `${data.name}.env.${env.key}`,
          value: `"${env.value}"`,
        })),
      ];
    }
    if (data.type === "artifact") {
      return [
        {
          id: `artifact.${id}`,
          displayName: `${data.name}`,
          value: `${normaliseNameToStarlarkVariable(data.name)}`,
        },
      ];
    }

    if (data.type === "shell") {
      return [
        {
          id: `shell.${id}`,
          displayName: `${data.name}`,
          value: `${normaliseNameToStarlarkVariable(data.name)}.files_artifacts[0]`,
        },
        ...data.env.map((env, i) => ({
          id: `shell.${id}.env.${i}.value`,
          displayName: `${data.name}.env.${env.key}`,
          value: `"${env.value}"`,
        })),
      ];
    }

    if (data.type === "python") {
      return [
        {
          id: `python.${id}`,
          displayName: `${data.name}`,
          value: `${normaliseNameToStarlarkVariable(data.name)}.files_artifacts[0]`,
        },
        ...data.args.map((arg, i) => ({
          id: `python.${id}.args.${i}.arg`,
          displayName: `${data.name}.args[${i}]`,
          value: `"${arg.arg}"`,
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
      const nameMatches = data.name.match(variablePattern);
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
        const portMatches = port.name.match(variablePattern) || port.applicationProtocol.match(variablePattern);
        if (portMatches) {
          getDependenciesFor(id).add(portMatches[2]);
        }
      });
      data.files.forEach((file) => {
        const fileMatches = file.mountPoint.match(variablePattern) || file.name.match(variablePattern);
        if (fileMatches) {
          getDependenciesFor(id).add(fileMatches[2]);
        }
      });
    }
    if (data.type === "exec") {
      const serviceMatches = data.service.match(variablePattern);
      if (serviceMatches) {
        getDependenciesFor(id).add(serviceMatches[2]);
      }
      const commandMatches = data.command.match(variablePattern);
      if (commandMatches) {
        getDependenciesFor(id).add(commandMatches[2]);
      }
    }
    if (data.type === "shell") {
      const nameMatches = data.name.match(variablePattern);
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
        const fileMatches = file.mountPoint.match(variablePattern) || file.name.match(variablePattern);
        if (fileMatches) {
          getDependenciesFor(id).add(fileMatches[2]);
        }
      });
    }
    if (data.type === "python") {
      const nameMatches = data.name.match(variablePattern);
      if (nameMatches) {
        getDependenciesFor(id).add(nameMatches[2]);
      }
      data.args.forEach((arg) => {
        const argMatches = arg.arg.match(variablePattern);
        if (argMatches) {
          getDependenciesFor(id).add(argMatches[2]);
        }
      });
      data.files.forEach((file) => {
        const fileMatches = file.mountPoint.match(variablePattern) || file.name.match(variablePattern);
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
  const nodeLookup = nodes.reduce((acc: Record<string, Node>, cur) => ({ ...acc, [cur.id]: cur }), {});
  const primaryNodes = nodes.filter((node) => !isDefined(node.parentNode));
  const primaryEdges = edges
    .map((edge) => ({ ...edge, source: nodeLookup[edge.source].parentNode || edge.source }))
    .filter((e) => e.target !== e.source);

  // Topological sort
  const sortedNodes: Node<KurtosisServiceNodeData>[] = [];
  let remainingEdges = [...primaryEdges];
  while (remainingEdges.length > 0 || sortedNodes.length !== primaryNodes.length) {
    const nodesToRemove = primaryNodes
      .filter((node) => remainingEdges.every((edge) => edge.target !== node.id)) // eslint-disable-line no-loop-func
      .filter((node) => !sortedNodes.includes(node));

    if (nodesToRemove.length === 0) {
      throw new Error(
        "Topological sort couldn't remove nodes. This indicates a cycle has been detected. Starlark cannot be rendered.",
      );
    }
    sortedNodes.push(...nodesToRemove);
    remainingEdges = remainingEdges.filter((edge) => sortedNodes.every((node) => edge.source !== node.id));
  }
  const variablesById = getVariablesFromNodes(data).reduce(
    (acc, cur) => ({ ...acc, [cur.id]: cur }),
    {} as Record<string, Variable>,
  );
  const interpolateValue = (input: string): string => {
    let formatString = input.replaceAll('"', '\\"');
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

  function objectToStarlark(o: any, indent: number) {
    const padLeft = "".padStart(indent, " ");
    if (!isDefined(o)) {
      return "None";
    }
    if (Array.isArray(o)) {
      let result = `[`;
      o.forEach((arrayValue) => {
        result += `${objectToStarlark(arrayValue, indent + 4)},\n`;
      });
      result += `${padLeft}],\n`;
      return result;
    }
    if (typeof o === "number") {
      return `${o}`;
    }
    if (typeof o === "string") {
      return interpolateValue(o);
    }
    if (typeof o === "boolean") {
      return o ? "True" : "False";
    }
    if (typeof o === "object") {
      let result = "{";
      Object.entries(o).forEach(([key, value]) => {
        result += `\n${padLeft}${interpolateValue(key)}: ${objectToStarlark(value, indent + 4)},`;
      });
      result += `${padLeft}}`;
      return result;
    }

    throw new Error(`Unable to convert the object ${o} to starlark`);
  }

  const renderImageConfig = (config: KurtosisImageConfig): string => {
    switch (config.type) {
      case "image":
        if ([config.registryUsername, config.registryPassword, config.registry].every((v) => v === "")) {
          return interpolateValue(config.image);
        }
        return `ImageSpec(name=${interpolateValue(config.image)}, username=${interpolateValue(
          config.registryUsername,
        )}, password=${config.registryPassword}, registry=${interpolateValue(config.registry)})`;
      case "dockerfile":
        return `ImageBuildSpec(image_name=${interpolateValue(config.image)}, build_context_dir=${interpolateValue(
          config.buildContextDir,
        )}, target_stage=${interpolateValue(config.targetStage)})`;
      case "nix":
        return `NixBuildSpec(image_name=${interpolateValue(config.image)}, build_context_dir=${interpolateValue(
          config.buildContextDir,
        )}, flake_location_dir=${interpolateValue(config.flakeLocationDir)}, flake_output=${interpolateValue(
          config.flakeOutput,
        )})`;
    }
  };

  let starlark = "";
  const packageNodeData = sortedNodes
    .map((n) => data[n.id])
    .filter((d) => d.type === "package") as KurtosisPackageNodeData[];
  for (const nodeData of packageNodeData) {
    const module_name = `${normaliseNameToStarlarkVariable(nodeData.name)}_module`;
    // Todo handle other paths
    starlark += `${module_name} = import_module(${interpolateValue(nodeData.packageId + "/main.star")})\n`;
  }
  if (packageNodeData.length > 0) {
    starlark += "\n";
  }

  starlark += "def run(plan):\n";
  for (const node of sortedNodes) {
    const nodeData = data[node.id];
    if (nodeData.type === "service") {
      const serviceName = normaliseNameToStarlarkVariable(nodeData.name);
      starlark += `    ${serviceName} = plan.add_service(\n`;
      starlark += `        name = ${interpolateValue(nodeData.name)},\n`;
      starlark += `        config = ServiceConfig (\n`;
      starlark += `            image = ${renderImageConfig(nodeData.image)},\n`;
      starlark += `            ports = {\n`;
      for (const { name, port, applicationProtocol, transportProtocol } of nodeData.ports) {
        starlark += `                ${interpolateValue(name)}: PortSpec(\n`;
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
      for (const { mountPoint, name } of nodeData.files) {
        starlark += `                ${interpolateValue(mountPoint)}: ${interpolateValue(name)},\n`;
      }
      starlark += `            },\n`;
      starlark += `        ),\n`;
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "exec") {
      const serviceName = normaliseNameToStarlarkVariable(interpolateValue(nodeData.service).replace(/\.name$/, ""));
      const execName = `${serviceName}_exec`;
      starlark += `    ${execName} = plan.exec(\n`;
      starlark += `        service_name = ${interpolateValue(nodeData.service)},\n`;
      starlark += `        recipe = ExecRecipe(\n`;
      starlark += `            command = [${nodeData.command.split(" ").map(interpolateValue).join(", ")}],`;
      starlark += `        ),\n`;
      if (nodeData.acceptableCodes.length > 0) {
        starlark += `        acceptable_codes = [${nodeData.acceptableCodes.map(({ value }) => value).join(", ")}],\n`;
      }
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "artifact") {
      const artifactName = normaliseNameToStarlarkVariable(nodeData.name);
      starlark += `    ${artifactName} = plan.render_templates(\n`;
      starlark += `        name = "${nodeData.name}",\n`;
      starlark += `        config = {\n`;
      for (const [fileName, fileText] of Object.entries(nodeData.files)) {
        starlark += `            "${fileName}": struct(\n`;
        starlark += `                template = """${escapeTemplateString(fileText)}""",\n`;
        starlark += `                data={},\n`;
        starlark += `            ),\n`;
      }
      starlark += `        },\n`;
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "shell") {
      const shellName = normaliseNameToStarlarkVariable(nodeData.name);
      starlark += `    ${shellName} = plan.run_sh(\n`;
      starlark += `        run = """${escapeString(nodeData.command)}""",\n`;
      const image = renderImageConfig(nodeData.image);
      if (image !== '""') {
        starlark += `        image = ${image},\n`;
      }
      starlark += `        env_vars = {\n`;
      for (const { key, value } of nodeData.env) {
        starlark += `            ${interpolateValue(key)}: ${interpolateValue(value)},\n`;
      }
      starlark += `        },\n`;
      starlark += `        files = {\n`;
      for (const { mountPoint, name } of nodeData.files) {
        starlark += `            ${interpolateValue(mountPoint)}: ${interpolateValue(name)},\n`;
      }
      starlark += `        },\n`;
      starlark += `        store = [\n`;
      starlark += `            StoreSpec(src = ${interpolateValue(nodeData.store)}, name="${shellName}"),\n`;
      starlark += `        ],\n`;
      const wait = interpolateValue(nodeData.wait);
      if (nodeData.wait_enabled === "false" || wait !== '""') {
        starlark += `        wait=${nodeData.wait_enabled === "true" ? wait : "None"},\n`;
      }
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "python") {
      const pythonName = normaliseNameToStarlarkVariable(nodeData.name);
      starlark += `    ${pythonName} = plan.run_python(\n`;
      starlark += `        run = """${escapeString(nodeData.command)}""",\n`;
      const image = renderImageConfig(nodeData.image);
      if (image !== '""') {
        starlark += `        image = ${image},\n`;
      }
      starlark += `        packages = [\n`;
      for (const { packageName } of nodeData.packages) {
        starlark += `            ${interpolateValue(packageName)},\n`;
      }
      starlark += `        ],\n`;
      starlark += `        args = [\n`;
      for (const { arg } of nodeData.args) {
        starlark += `            ${interpolateValue(arg)},\n`;
      }
      starlark += `        ],\n`;
      starlark += `        files = {\n`;
      for (const { mountPoint, name } of nodeData.files) {
        starlark += `            ${interpolateValue(mountPoint)}: ${interpolateValue(name)},\n`;
      }
      starlark += `        },\n`;
      if (nodeData.store !== "") {
        starlark += `        store = [\n`;
        starlark += `            StoreSpec(src = ${interpolateValue(nodeData.store)}, name="${pythonName}"),\n`;
        starlark += `        ],\n`;
      }
      const wait = interpolateValue(nodeData.wait);
      if (nodeData.wait_enabled === "false" || wait !== '""') {
        starlark += `        wait=${nodeData.wait_enabled === "true" ? wait : "None"},\n`;
      }
      starlark += `    )\n\n`;
    }

    if (nodeData.type === "package") {
      const packageName = normaliseNameToStarlarkVariable(nodeData.name);
      starlark += `    ${packageName} = ${packageName}_module.run(plan, **${objectToStarlark(nodeData.args, 8)}`;
      starlark += `    )\n\n`;
    }
  }

  // Delete any services from any existing enclave that aren't defined anymore
  if (isDefined(existingEnclave) && existingEnclave.services?.isOk) {
    for (const existingService of Object.values(existingEnclave.services.value.serviceInfo)) {
      const serviceNoLongerExists = nodes.every((node) => {
        const nodeData = data[node.id];
        return !isDefined(nodeData) || nodeData.type !== "service" || nodeData.name !== existingService.name;
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
