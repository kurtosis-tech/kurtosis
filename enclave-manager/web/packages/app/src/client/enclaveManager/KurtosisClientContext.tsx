import { Flex, Heading, Spinner } from "@chakra-ui/react";
import { assertDefined, isDefined, KurtosisAlert, stringifyError } from "kurtosis-ui-components";
import { createContext, PropsWithChildren, useContext, useEffect, useMemo, useState } from "react";
import { jwtToken } from "../../cookies";
import { KURTOSIS_CLOUD_EM_PAGE, KURTOSIS_CLOUD_EM_URL } from "../constants";
import { AuthenticatedKurtosisClient } from "./AuthenticatedKurtosisClient";
import { KurtosisClient } from "./KurtosisClient";
import { LocalKurtosisClient } from "./LocalKurtosisClient";

type KurtosisClientContextState = {
  client: KurtosisClient | null;
};

const KurtosisClientContext = createContext<KurtosisClientContextState>({ client: null });

export const KurtosisClientProvider = ({ children }: PropsWithChildren) => {
  const [client, setClient] = useState<KurtosisClient>();
  const [error, setError] = useState<string>();

  const errorHandlingClient = useMemo(() => {
    if (isDefined(client)) {
      return new Proxy(client, {
        get(target, prop: string | symbol) {
          if (
            prop === "getEnclaves" ||
            prop === "getServices" ||
            prop === "getStarlarkRun" ||
            prop === "listFilesArtifactNamesAndUuids"
          ) {
            return new Proxy(target[prop], {
              apply: (target, thisArg, argumentsList) => {
                const methodResult = Reflect.apply(target, thisArg, argumentsList) as ReturnType<typeof target>;
                return methodResult.then((r) => {
                  if (r.isErr) {
                    console.error(r.error);
                  }
                  return r;
                });
              },
            });
          } else {
            return Reflect.get(target, prop);
          }
        },
      });
    }
    return undefined;
  }, [client]);

  useEffect(() => {
    (async () => {
      // If the pathname starts with /gateway` then we are trying to use an Authenticated client.
      const path = window.location.pathname;

      try {
        setError(undefined);
        let newClient: KurtosisClient | null = null;

        if (path.startsWith("/gateway")) {
          const pathConfigPattern = /\/gateway\/ips\/([^/]+)\/ports\/([^/]+)(\/|$)/;
          const matches = path.match(pathConfigPattern);
          if (!matches) {
            throw Error(`Cannot configure an authenticated kurtosis client on this path: \`${path}\``);
          }

          const gatewayHost = matches[1];
          const port = parseInt(matches[2]);
          if (isNaN(port)) {
            throw Error(`Port ${port} is not a number.`);
          }

          if (isDefined(jwtToken)) {
            newClient = new AuthenticatedKurtosisClient(
              `${gatewayHost}`,
              jwtToken,
              new URL(`${window.location.protocol}//${window.location.host}/${KURTOSIS_CLOUD_EM_PAGE}`),
              new URL(`${window.location.protocol}//${window.location.host}${matches[0]}`),
            );
          } else {
            window.location.href = KURTOSIS_CLOUD_EM_URL;
          }
        } else {
          newClient = new LocalKurtosisClient();
        }

        if (isDefined(newClient)) {
          const checkResp = await newClient.checkHealth();
          if (checkResp.isErr) {
            setError("Cannot reach the enclave manager backend - is the Enclave Manager API running and accessible?");
            return;
          }
          setClient(newClient);
        }
      } catch (e: any) {
        console.error(e);
        setError(stringifyError(e));
      }
    })();
  }, []);

  if (errorHandlingClient) {
    return (
      <KurtosisClientContext.Provider value={{ client: errorHandlingClient }}>
        {children}
      </KurtosisClientContext.Provider>
    );
  } else {
    return (
      <Flex width="100%" direction="column" alignItems={"center"} gap={"1rem"} padding={"3rem"}>
        {!isDefined(error) && (
          <>
            <Spinner size={"xl"} />
            <Heading as={"h2"} fontSize={"2xl"}>
              Connecting to enclave manager...
            </Heading>
          </>
        )}
        {isDefined(error) && <KurtosisAlert message={error} />}
      </Flex>
    );
  }
};

export const useKurtosisClient = (): KurtosisClient => {
  const { client } = useContext(KurtosisClientContext);

  assertDefined(client, `useKurtosisClient used incorrectly - KurtosisClient is not currently available.`);

  return client;
};
