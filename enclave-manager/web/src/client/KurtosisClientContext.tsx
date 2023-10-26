import { Alert, AlertDescription, AlertIcon, AlertTitle, Flex, Heading, Spinner, useToast } from "@chakra-ui/react";
import { createContext, PropsWithChildren, useContext, useEffect, useMemo, useState } from "react";
import { assertDefined, isDefined, isStringTrue, sleep, stringifyError } from "../utils";
import { AuthenticatedKurtosisClient } from "./AuthenticatedKurtosisClient";
import { KurtosisClient } from "./KurtosisClient";
import { LocalKurtosisClient } from "./LocalKurtosisClient";

type KurtosisClientContextState = {
  client: KurtosisClient | null;
};

let kurtosisClientCache: KurtosisClient | undefined;

const KurtosisClientContext = createContext<KurtosisClientContextState>({ client: null });

export const KurtosisClientProvider = ({ children }: PropsWithChildren) => {
  const toast = useToast();
  const [client, setClient] = useState<KurtosisClient>();
  const [jwtToken, setJwtToken] = useState<string>();
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
                    toast({
                      title: "Error",
                      description: r.error.message,
                      status: "error",
                      position: "top",
                      variant: "solid",
                    });
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
  }, [client, toast]);
  kurtosisClientCache = errorHandlingClient;

  useEffect(() => {
    const receiveMessage = (event: MessageEvent) => {
      const message = event.data.message;
      switch (message) {
        case "jwtToken":
          const value = event.data.value;
          if (isDefined(value)) {
            setJwtToken(value);
          }
          break;
      }
    };
    window.addEventListener("message", receiveMessage);
    return () => window.removeEventListener("message", receiveMessage);
  }, []);

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const requireAuth = isStringTrue(searchParams.get("require_authentication"));
    const requestedApiHost = searchParams.get("api_host");
    // eslint-disable-next-line
    const preloadedPackage = searchParams.get("package");
    try {
      setError(undefined);
      if (requireAuth) {
        assertDefined(requestedApiHost, `The parameter 'requestedApiHost' is not defined`);
        if (isDefined(jwtToken)) {
          setClient(new AuthenticatedKurtosisClient(requestedApiHost, jwtToken));
        }
      } else {
        setClient(new LocalKurtosisClient());
      }
    } catch (e: any) {
      console.error(e);
      setError(stringifyError(e));
    }
  }, [jwtToken]);

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
        {isDefined(error) && (
          <Alert status="error">
            <AlertIcon />
            <AlertTitle>Error:</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
      </Flex>
    );
  }
};

export const useKurtosisClient = (): KurtosisClient => {
  const { client } = useContext(KurtosisClientContext);

  assertDefined(client, `useKurtosisClient used incorrectly - KurtosisClient is not currently available.`);

  return client;
};

export const getKurtosisClient = async (): Promise<KurtosisClient> => {
  let attempts = 0;
  while (attempts < 100) {
    if (isDefined(kurtosisClientCache)) {
      return kurtosisClientCache;
    }
    // This is required as the react-router can attempt to load data with its loader
    // function before we have determined which KurtosisClient
    // to use.
    await sleep(100);
    attempts += 1;
  }
  throw new Error("The Kurtosis Client never became available to getKurtosisClient.");
};
