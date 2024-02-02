import { isDefined, RemoveFunctions, stringifyError } from "kurtosis-ui-components";
import { Edge, Node } from "reactflow";
import { Result } from "true-myth";
import { EnclaveFullInfo } from "../../../types";
import { KurtosisServiceNodeData } from "./KurtosisServiceNode";

export const EMUI_BUILD_STATE_KEY = "EMUI_BUILD_STATE";

export function starlarkScriptContainsEMUIBuildState(script: string) {
  return script.includes(EMUI_BUILD_STATE_KEY);
}

export function getInitialGraphStateFromEnclave<T extends object>(
  enclave?: RemoveFunctions<EnclaveFullInfo>,
): Result<{ nodes: Node<T>[]; edges: Edge<any>[] }, string> {
  if (!isDefined(enclave)) {
    return Result.ok({ nodes: [], edges: [] });
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

export function generateStarlarkFromGraph(
  nodes: Node<KurtosisServiceNodeData>[],
  edges: Edge[],
  existingEnclave?: RemoveFunctions<EnclaveFullInfo>,
): string {
  let starlark = "def run(plan):\n";

  for (const node of nodes) {
    const serviceName = node.data.name.replace(/\s/g, "").toLowerCase();
    starlark += `    ${serviceName} = plan.add_service(\n`;
    starlark += `        name = "${node.data.name}",\n`;
    starlark += `        config = ServiceConfig (\n`;
    starlark += `            image = "${node.data.image}"\n`;
    starlark += `        ),\n`;
    starlark += `    )\n\n`;
  }

  // Delete any services from any existing enclave that aren't defined anymore
  if (isDefined(existingEnclave) && existingEnclave.services?.isOk) {
    for (const existingService of Object.values(existingEnclave.services.value.serviceInfo)) {
      if (!nodes.some((node) => node.data.name === existingService.name)) {
        starlark += `    plan.remove_service(name = "${existingService.name}")\n`;
      }
    }
  }

  starlark += `\n\n# ${EMUI_BUILD_STATE_KEY}=${btoa(JSON.stringify({ nodes, edges }))}`;

  return starlark;
}
