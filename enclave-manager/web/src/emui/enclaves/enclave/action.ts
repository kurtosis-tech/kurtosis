import { StarlarkRunResponseLine } from "enclave-manager-sdk/build/api_container_service_pb";
import { EnclaveInfo } from "enclave-manager-sdk/build/engine_service_pb";
import { ActionFunction, ActionFunctionArgs } from "react-router-dom";
import { KurtosisClient } from "../../../client/enclaveManager/KurtosisClient";
import { ConfigureEnclaveForm } from "../../../components/enclaves/configuration/types";
import { RemoveFunctions } from "../../../utils/types";

const handleEnclaveAction = async (
  kurtosisClient: KurtosisClient,
  { params, request }: ActionFunctionArgs,
): Promise<{ logs: AsyncIterable<StarlarkRunResponseLine> }> => {
  const { config, enclave, packageId } = (await request.json()) as {
    config: ConfigureEnclaveForm;
    packageId: string;
    enclave: RemoveFunctions<EnclaveInfo>;
  };

  const logs = await kurtosisClient.runStarlarkPackage(enclave, packageId, config.args);
  return { logs };
};

export const enclaveAction =
  (kurtosisClient: KurtosisClient): ActionFunction =>
  async (args) => {
    return handleEnclaveAction(kurtosisClient, args);
  };

export type EnclaveActionResolvedType = Awaited<ReturnType<typeof handleEnclaveAction>>;
