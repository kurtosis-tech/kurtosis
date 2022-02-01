import {EngineServiceClient} from "../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb";
import * as grpc from "@grpc/grpc-js";
import {err, ok, Result} from "neverthrow";
import log from "loglevel"
import {newCleanArgs, newCreateEnclaveArgs, newDestroyEnclaveArgs, newStopEnclaveArgs} from "../constructor_calls";
import {
    CleanArgs,
    CleanResponse,
    CreateEnclaveArgs,
    CreateEnclaveResponse,
    DestroyEnclaveArgs,
    EnclaveAPIContainerHostMachineInfo,
    EnclaveAPIContainerInfo,
    EnclaveAPIContainerStatus,
    EnclaveContainersStatus,
    EnclaveInfo,
    GetEnclavesResponse,
    GetEngineInfoResponse,
    StopEnclaveArgs
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as jspb from "google-protobuf";
import {ApiContainerServiceClient, EnclaveContext, EnclaveID} from "kurtosis-core-api-lib";
import {Status} from "@grpc/grpc-js/build/src/constants";
import * as semver from "semver"
import {KURTOSIS_ENGINE_VERSION} from "../../kurtosis_engine_version/kurtosis_engine_version";

const LOCAL_HOST_IP_ADDRESS_STR: string = "0.0.0.0";

const SHOULD_PUBLISH_ALL_PORTS: boolean = true;

const API_CONTAINER_LOG_LEVEL: string = "info";

export const DEFAULT_KURTOSIS_ENGINE_SERVER_PORT_NUM: number = 9710;

// Blank tells the engine server to use the default
const DEFAULT_API_CONTAINER_VERSION_TAG = "";

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
export class KurtosisContext {
    private readonly client: EngineServiceClient;

    private constructor(client: EngineServiceClient) {
        this.client = client;
    }

    // Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
    public static async newKurtosisContextFromLocalEngine(): Promise<Result<KurtosisContext, Error>>{
        const kurtosisEngineSocketStr: string = `${LOCAL_HOST_IP_ADDRESS_STR}:${DEFAULT_KURTOSIS_ENGINE_SERVER_PORT_NUM}`;

        let engineServiceClient: EngineServiceClient;
        // TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
        try {
            engineServiceClient = new EngineServiceClient(kurtosisEngineSocketStr, grpc.credentials.createInsecure());
        } catch(exception) {
            if (exception instanceof Error) {
                return err(exception);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the engine client that wasn't an error: " + exception
            ));
        }

        const getEngineInfoPromise: Promise<Result<GetEngineInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
            engineServiceClient.getEngineInfo(emptyArg, (error: grpc.ServiceError | null, response?: GetEngineInfoResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never " + "happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    if(error.code === Status.UNAVAILABLE){
                        resolve(err(new Error("The Kurtosis Engine Server is unavailable and is probably not running; you " +
                            "will need to start it using the Kurtosis CLI before you can create a connection to it")));
                    }
                    resolve(err(error));
                }
            })
        });

        const getEngineInfoResult: Result<GetEngineInfoResponse, Error> = await getEngineInfoPromise;
        if (!getEngineInfoResult.isOk()) {
            return err(getEngineInfoResult.error)
        }

        const engineInfoResponse: GetEngineInfoResponse = getEngineInfoResult.value;

        const runningEngineVersionStr: string = engineInfoResponse.getEngineVersion()

        const runningEngineSemver: semver.SemVer | null = semver.parse(runningEngineVersionStr)
        if (runningEngineSemver === null){
            log.warn(`We expected the running engine version to match format X.Y.Z, but instead got '${runningEngineVersionStr}'; this means that we can't verify the API library and engine versions match so you may encounter runtime errors`)
        }
      
        const libraryEngineSemver: semver.SemVer | null = semver.parse(KURTOSIS_ENGINE_VERSION)
        if (libraryEngineSemver === null){
            log.warn(`We expected the API library version to match format X.Y.Z, but instead got '${KURTOSIS_ENGINE_VERSION}'; this means that we can't verify the API library and engine versions match so you may encounter runtime errors`)
        }
       
        if(runningEngineSemver && libraryEngineSemver){
            const runningEngineMajorVersion = semver.major(runningEngineSemver)
            const runningEngineMinorVersion = semver.minor(runningEngineSemver)
            
            const libraryEngineMajorVersion = semver.major(libraryEngineSemver)
            const libraryEngineMinorVersion = semver.minor(libraryEngineSemver)

            const doApiVersionsMatch: boolean = libraryEngineMajorVersion == runningEngineMajorVersion && libraryEngineMinorVersion == runningEngineMinorVersion
            if (!doApiVersionsMatch) {
                return err(new Error(
                    `An API version mismatch was detected between the running engine version ${runningEngineSemver.version} and the engine version the library expects, ${libraryEngineSemver.version}; you should use the version of this library that corresponds to the running engine version`
                    ));
                }
        }

        const kurtosisContext: KurtosisContext = new KurtosisContext(engineServiceClient);

        return ok(kurtosisContext);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async createEnclave(
        enclaveId: string,
        isPartitioningEnabled: boolean,
    ): Promise<Result<EnclaveContext, Error>> {

        const args: CreateEnclaveArgs = newCreateEnclaveArgs(
            enclaveId,
            DEFAULT_API_CONTAINER_VERSION_TAG,
            API_CONTAINER_LOG_LEVEL,
            isPartitioningEnabled,
            SHOULD_PUBLISH_ALL_PORTS,
        );

        const createEnclavePromise: Promise<Result<CreateEnclaveResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.createEnclave(args, (error: grpc.ServiceError | null, response?: CreateEnclaveResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    resolve(err(error));
                }
            })
        });

        const createEnclaveResult: Result<CreateEnclaveResponse, Error> = await createEnclavePromise;
        if (!createEnclaveResult.isOk()) {
            return err(createEnclaveResult.error)
        }

        const response: CreateEnclaveResponse = createEnclaveResult.value;

        const enclaveInfo: EnclaveInfo | undefined = response.getEnclaveInfo();
        if (enclaveInfo === undefined) {
            return err(new Error("An error occurred creating enclave with ID " + enclaveId + " enclaveInfo is undefined; " +
                "this is a bug on this library" ))
        }
        const newEnclaveContextResult: Result<EnclaveContext, Error> = KurtosisContext.newEnclaveContextFromEnclaveInfo(enclaveInfo);
        if (newEnclaveContextResult.isErr()) {
            return err(new Error(`An error occurred creating an enclave context from a newly-created enclave; this should never happen`))
        }

        return ok(newEnclaveContextResult.value);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async getEnclaveContext(enclaveId: EnclaveID): Promise<Result<EnclaveContext, Error>> {

        const getEnclavesResponsePromise: Promise<Result<GetEnclavesResponse, Error>> = this.getEnclaveResponse();
        const getEnclavesResponseResult: Result<GetEnclavesResponse, Error> = await getEnclavesResponsePromise;
        if (!getEnclavesResponseResult.isOk()) {
            return err(getEnclavesResponseResult.error);
        }
        const getEnclavesResponse: GetEnclavesResponse = getEnclavesResponseResult.value;

        const allEnclaveInfo: jspb.Map<string, EnclaveInfo> = getEnclavesResponse.getEnclaveInfoMap()
        const maybeEnclaveInfo: EnclaveInfo | undefined = allEnclaveInfo.get(enclaveId);
        if (maybeEnclaveInfo === undefined) {
            return err(new Error(`No enclave with ID '${enclaveId}' found`))
        }
        const newEnclaveCtxResult: Result<EnclaveContext, Error> = KurtosisContext.newEnclaveContextFromEnclaveInfo(maybeEnclaveInfo);
        if (newEnclaveCtxResult.isErr()) {
            return err(newEnclaveCtxResult.error);
        }
        return ok(newEnclaveCtxResult.value);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async getEnclaves(): Promise<Result<Set<EnclaveID>, Error>>{

        const getEnclavesResponsePromise: Promise<Result<GetEnclavesResponse, Error>> = this.getEnclaveResponse();
        const getEnclavesResponseResult: Result<GetEnclavesResponse, Error> = await getEnclavesResponsePromise;
        if (!getEnclavesResponseResult.isOk()) {
            return err(getEnclavesResponseResult.error);
        }
        const getEnclavesResponse: GetEnclavesResponse = getEnclavesResponseResult.value;

        const result: Set<EnclaveID> = new Set();
        for (let enclaveId of getEnclavesResponse.getEnclaveInfoMap().keys()) {
            result.add(enclaveId);
        }
        return ok(result);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async stopEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>> {
        const args: StopEnclaveArgs = newStopEnclaveArgs(enclaveId)

        const stopEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.stopEnclave(args, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const stopEnclaveResult: Result<null, Error> = await stopEnclavePromise;
        if (!stopEnclaveResult.isOk()) {
            return err(stopEnclaveResult.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async destroyEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>> {
        const args: DestroyEnclaveArgs = newDestroyEnclaveArgs(enclaveId);

        const destroyEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.destroyEnclave(args, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const destroyEnclaveResult: Result<null, Error> = await destroyEnclavePromise;
        if (!destroyEnclaveResult.isOk()) {
            return err(destroyEnclaveResult.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
    public async clean( shouldCleanAll : boolean): Promise<Result<Set<string>, Error>>{

        const cleanArgs: CleanArgs = newCleanArgs(shouldCleanAll);

        const cleanPromise: Promise<Result<CleanResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.clean(cleanArgs, (error: grpc.ServiceError | null, response?: CleanResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this " +
                            "should never happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    resolve(err(error));
                }
            })
        });
        const cleanResult: Result<CleanResponse, Error> = await cleanPromise;
        if (!cleanResult.isOk()) {
            return err(cleanResult.error)
        }
        const cleanResponse: CleanResponse = cleanResult.value;

        const result: Set<string> = new Set();
        for (let enclaveID of cleanResponse.getRemovedEnclaveIdsMap().keys()) {
            result.add(enclaveID);
        }
        return ok(result);
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    private static newEnclaveContextFromEnclaveInfo(enclaveInfo: EnclaveInfo): Result<EnclaveContext, Error> {
        const enclaveContainersStatus = enclaveInfo.getContainersStatus()
        if (enclaveContainersStatus !== EnclaveContainersStatus.ENCLAVECONTAINERSSTATUS_RUNNING) {
            return err(new Error(`Enclave containers status was '${enclaveContainersStatus}', but we can't create an enclave context from a non-running enclave`))
        }

        const enclaveApiContainerStatus = enclaveInfo.getApiContainerStatus()
        if (enclaveApiContainerStatus !== EnclaveAPIContainerStatus.ENCLAVEAPICONTAINERSTATUS_RUNNING) {
            return err(new Error(`Enclave API container status was '${enclaveApiContainerStatus}', but we can't create an enclave context without a running API container`))
        }

        const apiContainerInfo: EnclaveAPIContainerInfo | undefined = enclaveInfo.getApiContainerInfo();
        if (apiContainerInfo === undefined) {
            return err(new Error(`API container was listed as running, but no API container info exists`))
        }
        const apiContainerHostMachineInfo: EnclaveAPIContainerHostMachineInfo | undefined = enclaveInfo.getApiContainerHostMachineInfo()
        if (apiContainerHostMachineInfo === undefined) {
            return err(new Error(`API container was listed as running, but no API container host machine info exists`))
        }

        const apiContainerHostMachineUrl: string = `${apiContainerHostMachineInfo.getIpOnHostMachine()}:${apiContainerHostMachineInfo.getPortOnHostMachine()}`

        let apiContainerClient: ApiContainerServiceClient;
        // TODO SECURITY: Use HTTPS!
        try {
            apiContainerClient = new ApiContainerServiceClient(apiContainerHostMachineUrl, grpc.credentials.createInsecure());
        } catch(exception) {
            if (exception instanceof Error) {
                return err(exception);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that" +
                " wasn't an error: " + exception
            ));
        }

        const result: EnclaveContext = new EnclaveContext(
            apiContainerClient,
            enclaveInfo.getEnclaveId(),
            enclaveInfo.getEnclaveDataDirpathOnHostMachine(),
        )
        return ok(result);
    }

    private async getEnclaveResponse(): Promise<Result<GetEnclavesResponse, Error>>{
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getEnclavesPromise: Promise<Result<GetEnclavesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEnclaves(emptyArg, (error: grpc.ServiceError | null, response?: GetEnclavesResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    resolve(err(error));
                }
            })
        });
        const getEnclavesResponseResult: Result<GetEnclavesResponse, Error> = await getEnclavesPromise;
        if (!getEnclavesResponseResult.isOk()) {
            return err(getEnclavesResponseResult.error)
        }

        return ok(getEnclavesResponseResult.value);
    }
}
