import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveAPIContainerInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { ActionFunction, ActionFunctionArgs } from "react-router-dom";
import { KurtosisClient } from "../../../client/enclaveManager/KurtosisClient";
import { ConfigureEnclaveForm } from "../../../components/enclaves/configuration/types";
import { RemoveFunctions } from "../../../utils/types";

const handleRunStarlarkAction = async (
  kurtosisClient: KurtosisClient,
  { params, request }: ActionFunctionArgs,
): Promise<{ logs: AsyncIterable<StarlarkRunResponseLine> }> => {
  const { config, apicInfo, packageId } = (await request.json()) as {
    config: ConfigureEnclaveForm;
    packageId: string;
    apicInfo: RemoveFunctions<EnclaveAPIContainerInfo>;
  };

  const logs = await kurtosisClient.runStarlarkPackage(apicInfo, packageId, config.args);
  return { logs };
};

export const runStarlarkAction =
  (kurtosisClient: KurtosisClient): ActionFunction =>
  async (args) => {
    return handleRunStarlarkAction(kurtosisClient, args);
  };

export type RunStarlarkResolvedType = Awaited<ReturnType<typeof handleRunStarlarkAction>>;
