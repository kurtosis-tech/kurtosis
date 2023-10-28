import { LoaderFunctionArgs } from "react-router-dom";
import { KurtosisClient } from "../../../client/enclaveManager/KurtosisClient";
import { isDefined } from "../../../utils";

export const enclaveLoader =
  (kurtosisClient: KurtosisClient) =>
  async ({ params }: LoaderFunctionArgs): Promise<{ routeName: string }> => {
    const uuid = params.enclaveUUID;

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

    return {
      routeName: enclave.name,
    };
  };
