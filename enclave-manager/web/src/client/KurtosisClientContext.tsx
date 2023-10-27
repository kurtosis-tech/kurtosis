import { Alert, AlertDescription, AlertIcon, AlertTitle, Flex, Heading, Spinner, useToast } from "@chakra-ui/react";
import { createContext, PropsWithChildren, useContext, useEffect, useMemo, useState } from "react";
import { assertDefined, isDefined, isStringTrue, stringifyError } from "../utils";
import { AuthenticatedKurtosisClient } from "./AuthenticatedKurtosisClient";
import { KurtosisClient } from "./KurtosisClient";
import { LocalKurtosisClient } from "./LocalKurtosisClient";

type KurtosisClientContextState = {
  client: KurtosisClient | null;
};

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
    (async () => {
      const searchParams = new URLSearchParams(window.location.search);
      const requireAuth = isStringTrue(searchParams.get("require_authentication"));
      const requestedApiHost = searchParams.get("api_host");
      // eslint-disable-next-line
      const preloadedPackage = searchParams.get("package");
      try {
        setError(undefined);
        let newClient: KurtosisClient | null = null;
        if (requireAuth) {
          assertDefined(requestedApiHost, `The parameter 'requestedApiHost' is not defined`);
          if (isDefined(jwtToken)) {
            newClient = new AuthenticatedKurtosisClient(requestedApiHost, jwtToken);
          }
        } else {
          newClient = new LocalKurtosisClient();
        }
        if (isDefined(newClient)) {
          const checkResp = await newClient.checkHealth();
          if (checkResp.isErr) {
            setError("Cannot reach the enclave manager backend - is your enclave manager definitely running?");
            return;
          }
          setClient(newClient);
        }
      } catch (e: any) {
        console.error(e);
        setError(stringifyError(e));
      }
    })();
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
