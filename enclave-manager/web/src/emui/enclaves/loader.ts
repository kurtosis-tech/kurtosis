import { defer } from "react-router-dom";
import { Result, ResultNS } from "true-myth";
import { KurtosisClient } from "../../client/enclaveManager/KurtosisClient";
import { EnclaveFullInfo } from "./types";

const loadEnclaves = async (kurtosisClient: KurtosisClient): Promise<Result<EnclaveFullInfo[], string>> => {
  const enclavesResponse = await kurtosisClient.getEnclaves();
  if (enclavesResponse.isErr) {
    return Result.err(enclavesResponse.error.message || "Unknown api error");
  }
  const enclaves = Object.values(enclavesResponse.value.enclaveInfo);
  const [starlarkRuns, services, filesAndArtifacts] = await Promise.all([
    Promise.all(enclaves.map((enclave) => kurtosisClient.getStarlarkRun(enclave))),
    Promise.all(enclaves.map((enclave) => kurtosisClient.getServices(enclave))),
    Promise.all(enclaves.map((enclave) => kurtosisClient.listFilesArtifactNamesAndUuids(enclave))),
  ]);

  const starlarkErrors = starlarkRuns.filter(ResultNS.isErr);
  const servicesErrors = services.filter(ResultNS.isErr);
  const filesAndArtifactErrors = filesAndArtifacts.filter(ResultNS.isErr);
  if (starlarkErrors.length + servicesErrors.length + filesAndArtifactErrors.length > 0) {
    return Result.err(
      `Starlark errors: ${
        starlarkErrors.length > 0 ? starlarkErrors.map((r) => r.error.message).join("\n") : "None"
      }\nServices errors: ${
        servicesErrors.length > 0 ? servicesErrors.map((r) => r.error.message).join("\n") : "None"
      }\nFiles and Artifacts errors: ${
        filesAndArtifactErrors.length > 0 ? filesAndArtifactErrors.map((r) => r.error.message).join("\n") : "None"
      }`,
    );
  }

  return Result.ok(
    enclaves.map((enclave, i) => ({
      ...enclave,
      // These values are never actually null because of the checking above
      starlarkRun: starlarkRuns[i].unwrapOr(null)!,
      services: services[i].unwrapOr(null)!,
      filesAndArtifacts: filesAndArtifacts[i].unwrapOr(null)!,
    })),
  );
};

export type EnclavesLoaderResolved = {
  enclaves: Awaited<ReturnType<typeof loadEnclaves>>;
};

export const enclavesLoader = (kurtosisClient: KurtosisClient) => async () => {
  return defer({ enclaves: loadEnclaves(kurtosisClient) });
};
