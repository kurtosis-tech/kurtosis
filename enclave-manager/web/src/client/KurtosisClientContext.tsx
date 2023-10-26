import { Alert, AlertDescription, AlertIcon, AlertTitle, Flex, Heading, Spinner } from "@chakra-ui/react";
import { createContext, PropsWithChildren, useContext, useEffect, useState } from "react";
import { assertDefined, isDefined, isStringTrue, stringifyError } from "../utils";
import { AuthenticatedKurtosisClient } from "./AuthenticatedKurtosisClient";
import { KurtosisClient } from "./KurtosisClient";
import { LocalKurtosisClient } from "./LocalKurtosisClient";

type KurtosisClientContextState = {
  client: KurtosisClient | null;
};

const KurtosisClientContext = createContext<KurtosisClientContextState>({ client: null });

export const KurtosisClientProvider = ({ children }: PropsWithChildren) => {
  const [client, setClient] = useState<KurtosisClient>();
  const [jwtToken, setJwtToken] = useState<string>();
  const [error, setError] = useState<string>();

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

  if (client) {
    return <KurtosisClientContext.Provider value={{ client }}>{children}</KurtosisClientContext.Provider>;
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
