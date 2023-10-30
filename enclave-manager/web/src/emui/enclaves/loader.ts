import { defer } from "react-router-dom";
import { Result } from "true-myth";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { EnclaveFullInfo } from "./types";

const loadEnclaves = async (kurtosisClient: KurtosisClient): Promise<Result<EnclaveFullInfo[], string>> => {
  const enclavesResponse = await kurtosisClient.getEnclaves();
  if (enclavesResponse.isErr) {
    return Result.err(enclavesResponse.error || "Unknown api error");
  }
  const enclaves = Object.values(enclavesResponse.value.enclaveInfo);
  const [starlarkRuns, services, filesAndArtifacts] = await Promise.all([
    Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
    Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
    Promise.all(enclaves.map((enclave) => kurtosisClient.listFilesArtifactNamesAndUuids(enclave))),
  ]);

  return Result.ok(
    enclaves.map((enclave, i) => ({
      ...enclave,
      // These values are never actually null because of the checking above
      starlarkRun: starlarkRuns[i],
      services: services[i],
      filesAndArtifacts: filesAndArtifacts[i],
    })),
  );
};

export type EnclavesLoaderResolved = {
  enclaves: Awaited<ReturnType<typeof loadEnclaves>>;
};

export const enclavesLoader = (kurtosisClient: KurtosisClient) => async () => {
  return defer({ enclaves: loadEnclaves(kurtosisClient) });
};
