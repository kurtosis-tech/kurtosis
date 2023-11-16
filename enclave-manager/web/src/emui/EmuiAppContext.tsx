import { Empty } from "@bufbuild/protobuf";
import { Flex, Heading, Spinner } from "@chakra-ui/react";
import {
  GetServicesResponse,
  GetStarlarkRunResponse,
  ListFilesArtifactNamesAndUuidsResponse,
  StarlarkRunResponseLine,
} from "enclave-manager-sdk/build/api_container_service_pb";
import {
  CreateEnclaveResponse,
  EnclaveAPIContainerInfo,
  EnclaveInfo,
} from "enclave-manager-sdk/build/engine_service_pb";
import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useState,
} from "react";
import { Result } from "true-myth";
import { useKurtosisClient } from "../client/enclaveManager/KurtosisClientContext";
import { isDefined } from "../utils";
import { RemoveFunctions } from "../utils/types";
import { EnclaveFullInfo } from "./enclaves/types";

export type EmuiAppState = {
  enclaves: Result<RemoveFunctions<EnclaveInfo>[], string>;
  servicesByEnclave: Record<string, Result<GetServicesResponse, string>>;
  filesAndArtifactsByEnclave: Record<string, Result<ListFilesArtifactNamesAndUuidsResponse, string>>;
  starlarkRunsByEnclave: Record<string, Result<GetStarlarkRunResponse, string>>;

  // Methods
  refreshEnclaves: () => Promise<Result<RemoveFunctions<EnclaveInfo>[], string>>;
  refreshServices: (enclave: RemoveFunctions<EnclaveInfo>) => Promise<Result<GetServicesResponse, string>>;
  refreshFilesAndArtifacts: (
    enclave: RemoveFunctions<EnclaveInfo>,
  ) => Promise<Result<ListFilesArtifactNamesAndUuidsResponse, string>>;
  refreshStarlarkRun: (enclave: RemoveFunctions<EnclaveInfo>) => Promise<Result<GetStarlarkRunResponse, string>>;
  createEnclave: (
    enclaveName: string,
    apiContainerLogLevel: string,
    productionMode?: boolean,
    apiContainerVersionTag?: string,
  ) => Promise<Result<CreateEnclaveResponse, string>>;
  destroyEnclave: (enclaveUUID: string) => Promise<Result<Empty, string>>;
  runStarlarkPackage: (
    apicInfo: RemoveFunctions<EnclaveAPIContainerInfo>,
    packageId: string,
    args: Record<string, any>,
  ) => Promise<AsyncIterable<StarlarkRunResponseLine>>;
};

const EmuiAppContext = createContext<EmuiAppState>({
  enclaves: Result.err("Enclaves not initialised, call refreshEnclaves"),
  servicesByEnclave: {},
  filesAndArtifactsByEnclave: {},
  starlarkRunsByEnclave: {},
  refreshEnclaves: () => null as any,
  refreshServices: () => null as any,
  refreshFilesAndArtifacts: () => null as any,
  refreshStarlarkRun: () => null as any,
  createEnclave: () => null as any,
  destroyEnclave: () => null as any,
  runStarlarkPackage: () => null as any,
});

export const EmuiAppContextProvider = ({ children }: PropsWithChildren) => {
  const [isInitialLoading, setIsInitialLoading] = useState(true);

  const [state, setState] = useState<RemoveFunctions<EmuiAppState>>({
    enclaves: Result.err("Enclaves not initialised, call refreshEnclaves"),
    servicesByEnclave: {},
    filesAndArtifactsByEnclave: {},
    starlarkRunsByEnclave: {},
  });
  const kurtosisClient = useKurtosisClient();

  const refreshEnclaves = useCallback(async () => {
    const getEnclavesResponse = await kurtosisClient.getEnclaves();
    setState((state) => ({
      ...state,
      enclaves: getEnclavesResponse.map((resp) => Object.values(resp.enclaveInfo)),
    }));
    return getEnclavesResponse.map((resp) => Object.values(resp.enclaveInfo));
  }, [kurtosisClient]);

  const refreshServices = useCallback(
    async (enclave: RemoveFunctions<EnclaveInfo>) => {
      const getServicesResponse = await kurtosisClient.getServices(enclave);
      setState((state) => ({
        ...state,
        servicesByEnclave: { ...state.servicesByEnclave, [enclave.shortenedUuid]: getServicesResponse },
      }));
      return getServicesResponse;
    },
    [kurtosisClient],
  );

  const refreshFilesAndArtifacts = useCallback(
    async (enclave: RemoveFunctions<EnclaveInfo>) => {
      const listFilesArtifactNamesAndUuidsResponse = await kurtosisClient.listFilesArtifactNamesAndUuids(enclave);
      setState((state) => ({
        ...state,
        filesAndArtifactsByEnclave: {
          ...state.filesAndArtifactsByEnclave,
          [enclave.shortenedUuid]: listFilesArtifactNamesAndUuidsResponse,
        },
      }));
      return listFilesArtifactNamesAndUuidsResponse;
    },
    [kurtosisClient],
  );

  const refreshStarlarkRun = useCallback(
    async (enclave: RemoveFunctions<EnclaveInfo>) => {
      const getStarlarkRunResponse = await kurtosisClient.getStarlarkRun(enclave);
      setState((state) => ({
        ...state,
        starlarkRunsByEnclave: { ...state.starlarkRunsByEnclave, [enclave.shortenedUuid]: getStarlarkRunResponse },
      }));
      return getStarlarkRunResponse;
    },
    [kurtosisClient],
  );

  const createEnclave = useCallback(
    async (
      enclaveName: string,
      apiContainerLogLevel: string,
      productionMode?: boolean,
      apiContainerVersionTag?: string,
    ) => {
      const resp = await kurtosisClient.createEnclave(
        enclaveName,
        apiContainerLogLevel,
        productionMode,
        apiContainerVersionTag,
      );
      if (resp.isOk && isDefined(resp.value.enclaveInfo)) {
        setState((state) => ({
          ...state,
          enclaves: state.enclaves.isOk
            ? Result.ok([...state.enclaves.value, resp.value.enclaveInfo].filter(isDefined))
            : state.enclaves,
        }));
      }
      return resp;
    },
    [kurtosisClient],
  );

  const destroyEnclave = useCallback(
    async (enclaveUUID: string) => {
      const resp = await kurtosisClient.destroy(enclaveUUID);
      if (resp.isOk) {
        setState((state) => ({
          ...state,
          enclaves: state.enclaves.isOk
            ? Result.ok(state.enclaves.value.filter((enclave) => enclave.enclaveUuid !== enclaveUUID))
            : state.enclaves,
        }));
      }
      return resp;
    },
    [kurtosisClient],
  );

  const runStarlarkPackage = useCallback(
    async (apicInfo: RemoveFunctions<EnclaveAPIContainerInfo>, packageId: string, args: Record<string, any>) => {
      const resp = await kurtosisClient.runStarlarkPackage(apicInfo, packageId, args);
      // TODO: Proxy lines to build optimistic ui
      return resp;
    },
    [kurtosisClient],
  );

  useEffect(() => {
    (async () => {
      await refreshEnclaves();
      setIsInitialLoading(false);
    })();
  }, [refreshEnclaves]);

  if (isInitialLoading) {
    return (
      <Flex width="100%" direction="column" alignItems={"center"} gap={"1rem"} padding={"3rem"}>
        <Spinner size={"xl"} />
        <Heading as={"h2"} fontSize={"2xl"}>
          Fetching Enclaves...
        </Heading>
      </Flex>
    );
  }

  return (
    <EmuiAppContext.Provider
      value={{
        ...state,
        refreshEnclaves,
        refreshStarlarkRun,
        refreshFilesAndArtifacts,
        refreshServices,
        createEnclave,
        destroyEnclave,
        runStarlarkPackage,
      }}
    >
      {children}
    </EmuiAppContext.Provider>
  );
};

export const useEmuiAppContext = () => {
  return useContext(EmuiAppContext);
};

export const useFullEnclave = (enclaveUUID: string): Result<EnclaveFullInfo, string> => {
  const {
    enclaves,
    servicesByEnclave,
    filesAndArtifactsByEnclave,
    starlarkRunsByEnclave,
    refreshServices,
    refreshStarlarkRun,
    refreshFilesAndArtifacts,
  } = useEmuiAppContext();

  const enclave = enclaves.isOk ? enclaves.value.find((enclave) => enclave.shortenedUuid === enclaveUUID) : null;

  const services = servicesByEnclave[enclaveUUID];
  const filesAndArtifacts = filesAndArtifactsByEnclave[enclaveUUID];
  const starlarkRun = starlarkRunsByEnclave[enclaveUUID];

  useEffect(() => {
    if (isDefined(enclave) && !isDefined(services)) {
      refreshServices(enclave);
    }
  }, [services, refreshServices, enclave]);

  useEffect(() => {
    if (isDefined(enclave) && !isDefined(filesAndArtifacts)) {
      refreshFilesAndArtifacts(enclave);
    }
  }, [filesAndArtifacts, refreshFilesAndArtifacts, enclave]);

  useEffect(() => {
    if (isDefined(enclave) && !isDefined(starlarkRun)) {
      refreshStarlarkRun(enclave);
    }
  }, [starlarkRun, refreshStarlarkRun, enclave]);

  if (enclaves.isErr) {
    return enclaves.cast();
  }

  if (!isDefined(enclave)) {
    return Result.err(`Could not find enclave ${enclaveUUID}`);
  }

  return Result.ok({
    ...enclave,
    services,
    filesAndArtifacts,
    starlarkRun,
  });
};

export const useFullEnclaves = (): Result<EnclaveFullInfo[], string> => {
  const {
    enclaves,
    servicesByEnclave,
    filesAndArtifactsByEnclave,
    starlarkRunsByEnclave,
    refreshServices,
    refreshStarlarkRun,
    refreshFilesAndArtifacts,
  } = useEmuiAppContext();

  // This hook can trigger a lot of requests to refresh data. To avoid creating waterfalls
  // of effects this refreshId along with cache values are used to restrict changes to the
  // useEffect dependency array.
  const [refreshId, incRefreshId] = useReducer((x: number) => x + 1, 0);
  const [cachedServicesByEnclave, cachedFilesAndArtifactsByEnclave, cachedStarlarkRunsByEnclave] = useMemo(
    () => [servicesByEnclave, filesAndArtifactsByEnclave, starlarkRunsByEnclave],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [refreshId],
  );

  useEffect(() => {
    if (enclaves.isOk) {
      (async () => {
        await Promise.all([
          ...enclaves.value
            .map((enclave) =>
              isDefined(cachedServicesByEnclave[enclave.shortenedUuid]) ? null : refreshServices(enclave),
            )
            .filter(isDefined),
          ...enclaves.value
            .map((enclave) =>
              isDefined(cachedFilesAndArtifactsByEnclave[enclave.shortenedUuid])
                ? null
                : refreshFilesAndArtifacts(enclave),
            )
            .filter(isDefined),
          ...enclaves.value
            .map((enclave) =>
              isDefined(cachedStarlarkRunsByEnclave[enclave.shortenedUuid]) ? null : refreshStarlarkRun(enclave),
            )
            .filter(isDefined),
        ]);
        incRefreshId();
      })();
    }
  }, [
    enclaves,
    refreshStarlarkRun,
    refreshServices,
    refreshFilesAndArtifacts,
    cachedFilesAndArtifactsByEnclave,
    cachedServicesByEnclave,
    cachedStarlarkRunsByEnclave,
  ]);

  const fullEnclaves = useMemo(
    () =>
      enclaves.map((enclaves) =>
        enclaves.map((enclave) => ({
          ...enclave,
          services: cachedServicesByEnclave[enclave.shortenedUuid],
          filesAndArtifacts: cachedFilesAndArtifactsByEnclave[enclave.shortenedUuid],
          starlarkRun: cachedStarlarkRunsByEnclave[enclave.shortenedUuid],
        })),
      ),
    [enclaves, cachedServicesByEnclave, cachedStarlarkRunsByEnclave, cachedFilesAndArtifactsByEnclave],
  );

  return fullEnclaves;
};
