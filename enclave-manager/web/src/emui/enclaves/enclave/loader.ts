import { defer, LoaderFunctionArgs } from "react-router-dom";

import { Result } from "true-myth";
import { KurtosisClient } from "../../../client/enclaveManager/KurtosisClient";
import { isDefined } from "../../../utils";
import { EnclaveFullInfo } from "../types";

export const loadEnclave = async (
  kurtosisClient: KurtosisClient,
  uuid?: string,
): Promise<{ routeName: string; enclave?: Result<EnclaveFullInfo, string> }> => {
  if (!isDefined(uuid)) {
    return {
      routeName: "Missing uuid",
    };
  }

  const enclavesResult = await kurtosisClient.getEnclaves();
  if (enclavesResult.isErr) {
    return {
      routeName: uuid,
    };
  }

  const enclave = Object.values(enclavesResult.value.enclaveInfo).find((enclave) => enclave.shortenedUuid === uuid);
  if (!isDefined(enclave)) {
    return {
      routeName: uuid,
    };
  }

  const [services, starlarkRun, filesAndArtifacts] = await Promise.all([
    kurtosisClient.getServices(enclave),
    kurtosisClient.getStarlarkRun(enclave),
    kurtosisClient.listFilesArtifactNamesAndUuids(enclave),
  ]);

  if (services.isErr) {
    return {
      routeName: enclave.name,
      enclave: Result.err(`Could not get services for enclave ${enclave.shortenedUuid}: ${services.error.message}`),
    };
  }

  if (starlarkRun.isErr) {
    return {
      routeName: enclave.name,
      enclave: Result.err(
        `Could not get starlark run for enclave ${enclave.shortenedUuid}: ${starlarkRun.error.message}`,
      ),
    };
  }

  if (filesAndArtifacts.isErr) {
    return {
      routeName: enclave.name,
      enclave: Result.err(
        `Could not get files for enclave ${enclave.shortenedUuid}: ${filesAndArtifacts.error.message}`,
      ),
    };
  }

  return {
    routeName: enclave.name,
    enclave: Result.ok({
      ...enclave,
      starlarkRun: starlarkRun.value,
      services: services.value,
      filesAndArtifacts: filesAndArtifacts.value,
    }),
  };
};

export const enclaveLoader =
  (kurtosisClient: KurtosisClient) =>
  ({ params }: LoaderFunctionArgs) => {
    return defer({ data: loadEnclave(kurtosisClient, params.enclaveUUID) });
  };

export type EnclaveLoaderDeferred = { data: ReturnType<typeof loadEnclave> };
export type EnclaveLoaderResolved = { data: Awaited<ReturnType<typeof loadEnclave>> };
