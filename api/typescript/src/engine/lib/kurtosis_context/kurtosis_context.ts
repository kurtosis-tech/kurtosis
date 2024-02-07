import log from "loglevel"
import * as semver from "semver"
import {err, ok, Result} from "neverthrow";
import {EnclaveContext, EnclaveUUID, ServiceUUID} from "../../../index";
import { GenericEngineClient } from "./generic_engine_client";
import { KURTOSIS_VERSION } from "../../../kurtosis_version/kurtosis_version";
import { GrpcNodeEngineClient } from "./grpc_node_engine_client";
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
    StopEnclaveArgs,
    GetServiceLogsArgs, EnclaveNameAndUuid,
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import {
    newCleanArgs,
    newCreateEnclaveArgs,
    newDestroyEnclaveArgs,
    newGetServiceLogsArgs,
    newStopEnclaveArgs
} from "../constructor_calls";
import {Readable} from "stream";
import {LogLineFilter} from "./log_line_filter";
import {Enclaves} from "./enclaves";
import {EnclaveIdentifiers} from "./enclave_identifiers";

const LOCAL_HOSTNAME: string = "localhost";

const API_CONTAINER_LOG_LEVEL: string = "debug";

const SHORTENED_UUID_ALLOWED_MATCHES = 1;

export const DEFAULT_GRPC_ENGINE_SERVER_PORT_NUM: number = 9710;

// Blank tells the engine server to use the default
const DEFAULT_API_CONTAINER_VERSION_TAG = "";

const DEFAULT_SHOULD_APIC_RUN_IN_DEBUG_MODE = false
const RUN_APIC_IN_DEBUG_MODE = true

// Docs available at https://docs.kurtosis.com/sdk#kurtosiscontext
export class KurtosisContext {
    private readonly client: GenericEngineClient

    constructor(client: GenericEngineClient){
        this.client = client;
    }

    // Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
    public static async newKurtosisContextFromLocalEngine():Promise<Result<KurtosisContext, Error>>  {
        let genericEngineClient: GenericEngineClient
        try {
            //These imports are dynamically imported here, otherwise compiling in Web environment fails for 2 reasons:
            // 1. "@grpc/grpc-js" could ONLY be run in Node environment(because of it's own dependencies). So importing it on top of the file will break compilation.
            // 2. WebPack compiler intents to check the libs no matter if those are behind IF statement. Which also break. That's why /* webpackIgnore: true */, avoid checkings.

            // 'engine_service_grpc_pb' has it's own "@grpc/grpc-js" import, that's why we import it dynamically also.

            const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")
            const engineServiceNode = await import( /* webpackIgnore: true */ "../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb")

            const kurtosisEngineSocketStr: string = `${LOCAL_HOSTNAME}:${DEFAULT_GRPC_ENGINE_SERVER_PORT_NUM}`
            const engineServiceClientNode = new engineServiceNode.EngineServiceClient(kurtosisEngineSocketStr, grpc_node.credentials.createInsecure())
            genericEngineClient = new GrpcNodeEngineClient(engineServiceClientNode)
        } catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the engine client that wasn't an error: " + error
            ));
        }

        const engineApiVersionValidationResult = await KurtosisContext.validateEngineApiVersion(genericEngineClient)
        if(engineApiVersionValidationResult.isErr()){
            return err(engineApiVersionValidationResult.error)
        }

        const kurtosisContext = new KurtosisContext(genericEngineClient)
        return ok(kurtosisContext)
    }

    // Docs available at https://docs.kurtosis.com/sdk#createenclaveenclaveid-enclaveid-boolean-issubnetworkingenabled---enclavecontextenclavecontext-enclavecontext
    public async createEnclave(enclaveName: string): Promise<Result<EnclaveContext, Error>> {
        const enclaveArgs: CreateEnclaveArgs = newCreateEnclaveArgs(
            enclaveName,
            DEFAULT_API_CONTAINER_VERSION_TAG,
            API_CONTAINER_LOG_LEVEL,
            DEFAULT_SHOULD_APIC_RUN_IN_DEBUG_MODE,
        );

        return this.createEnclaveWithArgs(enclaveArgs)
    }

    public async createEnclaveWithDebugEnabled(enclaveName: string): Promise<Result<EnclaveContext, Error>> {
        const enclaveArgs: CreateEnclaveArgs = newCreateEnclaveArgs(
            enclaveName,
            DEFAULT_API_CONTAINER_VERSION_TAG,
            API_CONTAINER_LOG_LEVEL,
            RUN_APIC_IN_DEBUG_MODE,
        );

        return this.createEnclaveWithArgs(enclaveArgs)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getenclavecontextstring-enclaveidentifier---enclavecontextenclavecontext-enclavecontext
    public async getEnclaveContext(enclaveIdentifier: string): Promise<Result<EnclaveContext, Error>> {
        const enclaveInfoResult = await this.getEnclave(enclaveIdentifier)

        if (enclaveInfoResult.isErr()) {
            return err(enclaveInfoResult.error)
        }

        const enclaveInfo = enclaveInfoResult.value

        const newEnclaveContextResult: Result<EnclaveContext, Error> = await this.newEnclaveContextFromEnclaveInfo(enclaveInfo);
        if (newEnclaveContextResult.isErr()) {
            return err(newEnclaveContextResult.error);
        }

        return ok(newEnclaveContextResult.value);
    }

    // Docs available at https://docs.kurtosis.com/sdk#getenclaves---enclaves-enclaves
    public async getEnclaves(): Promise<Result<Enclaves, Error>>{
        const getEnclavesResponseResult = await this.client.getEnclavesResponse();
        if (getEnclavesResponseResult.isErr()) {
            return err(getEnclavesResponseResult.error);
        }

        const getEnclavesResponse: GetEnclavesResponse = getEnclavesResponseResult.value;
        const enclavesByUuid : Map<EnclaveUUID, EnclaveInfo> = new Map<EnclaveUUID, EnclaveInfo>()
        const enclavesByName : Map<string, EnclaveInfo> = new Map<EnclaveUUID, EnclaveInfo>()
        const enclavesByShortenedUuid : Map<string, EnclaveInfo[]> = new Map<EnclaveUUID, EnclaveInfo[]>()
        getEnclavesResponse.getEnclaveInfoMap().forEach((enclaveInfo: EnclaveInfo, enclaveUuid :string) => {
            enclavesByUuid.set(enclaveUuid, enclaveInfo)
            enclavesByName.set(enclaveInfo.getName(), enclaveInfo)
            const shortenedUuid = enclaveInfo.getShortenedUuid()
            if (enclavesByShortenedUuid.has(shortenedUuid)) {
                enclavesByShortenedUuid.get(shortenedUuid)!.push(enclaveInfo)
            } else {
                enclavesByShortenedUuid.set(shortenedUuid,  [enclaveInfo])
            }
        });

        return ok(new Enclaves(enclavesByUuid, enclavesByName, enclavesByShortenedUuid));
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getenclavestring-enclaveidentifier---enclaveinfo-enclaveinfo
    public async getEnclave(enclaveIdentifier: string): Promise<Result<EnclaveInfo, Error>> {
        const enclavesResult = await this.getEnclaves()
        if (enclavesResult.isErr()) {
            return err(enclavesResult.error)
        }

        const enclaves = enclavesResult.value

        if (enclaves.enclavesByUuid.has(enclaveIdentifier)) {
            return ok(enclaves.enclavesByUuid.get(enclaveIdentifier)!)
        }

        if (enclaves.enclavesByShortenedUuid.has(enclaveIdentifier)) {
            const matchingEnclaves  = enclaves.enclavesByShortenedUuid.get(enclaveIdentifier)!
            if (matchingEnclaves.length == SHORTENED_UUID_ALLOWED_MATCHES) {
                return ok(matchingEnclaves[0])
            } else if (matchingEnclaves.length > SHORTENED_UUID_ALLOWED_MATCHES) {
                return err(new Error(`Found multiple ${matchingEnclaves} matches for shortened uuid ${enclaveIdentifier}`))
            }
        }

        if (enclaves.enclavesByName.has(enclaveIdentifier)) {
            return ok(enclaves.enclavesByName.get(enclaveIdentifier)!)
        }

        return err(new Error(`Couldn't find enclave for identifier '${enclaveIdentifier}'`))
    }

    // Docs available at https://docs.kurtosis.com/sdk/#stopenclavestring-enclaveidentifier
    public async stopEnclave(enclaveIdentifier: string): Promise<Result<null, Error>>{
        const stopEnclaveArgs: StopEnclaveArgs = newStopEnclaveArgs(enclaveIdentifier)
        const stopEnclaveResult = await this.client.stopEnclave(stopEnclaveArgs)
        if(stopEnclaveResult.isErr()){
            return err(stopEnclaveResult.error)
        }

        return ok(null)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#destroyenclavestring-enclaveidentifier
    public async destroyEnclave(enclaveIdentifier: string): Promise<Result<null, Error>>{
        const destroyEnclaveArgs: DestroyEnclaveArgs = newDestroyEnclaveArgs(enclaveIdentifier);
        const destroyEnclaveResult = await this.client.destroyEnclave(destroyEnclaveArgs)
        if(destroyEnclaveResult.isErr()){
            return err(destroyEnclaveResult.error)
        }

        return ok(null)
    }

    // Docs available at https://docs.kurtosis.com/sdk#cleanboolean-shouldcleanall---enclavenameanduuid-removedenclavenameanduuids
    public async clean(shouldCleanAll: boolean): Promise<Result<EnclaveNameAndUuid[], Error>>{
        const cleanArgs: CleanArgs = newCleanArgs(shouldCleanAll);
        const cleanResponseResult = await this.client.clean(cleanArgs)
        if(cleanResponseResult.isErr()){
            return err(cleanResponseResult.error)
        }

        const cleanResponse: CleanResponse = cleanResponseResult.value

        return ok(cleanResponse.getRemovedEnclaveNameAndUuidsList())
    }

    //The Readable object returned will be constantly streaming the service logs information using the ServiceLogsStreamContent
    //which contains two methods, the `getServiceLogsByServiceUuids` will return a map containing the service logs lines grouped by the service's UUID
    //and the `getNotFoundServiceUuids` will return set of not found (in the logs database) service UUIDs
    //Example of how to read the stream:
    //
    //serviceLogsReadable.on('data', (serviceLogsStreamContent: serviceLogsStreamContent) => {
    //      const serviceLogsByServiceUuids: Map<ServiceUUID, Array<ServiceLog>> = serviceLogsStreamContent.getServiceLogsByServiceUuids()
    //
    //      const notFoundServiceUuids: Set<ServiceUUID> = serviceLogsStreamContent.getNotFoundServiceUuids()
    //
    //      //insert your code here
    //})
    //You can cancel receiving the stream from the service calling serviceLogsReadable.destroy()
    // Docs available at https://docs.kurtosis.com/sdk#getservicelogsstring-enclaveidentifier-setserviceuuid-serviceuuids-boolean-shouldfollowlogs-loglinefilter-loglinefilter---servicelogsstreamcontent-servicelogsstreamcontent
    public async getServiceLogs(
        enclaveIdentifier: string,
        serviceUuids: Set<ServiceUUID>,
        shouldFollowLogs: boolean,
        shouldReturnAllLogs : boolean,
        numLogLines : number,
        logLineFilter: LogLineFilter|undefined): Promise<Result<Readable, Error>> {
        let getServiceLogsArgs: GetServiceLogsArgs;

        try {
            getServiceLogsArgs = newGetServiceLogsArgs(enclaveIdentifier, serviceUuids, shouldFollowLogs, shouldReturnAllLogs, numLogLines, logLineFilter);
        } catch(error) {
            return err(new Error(`An error occurred getting the get service logs arguments for enclave identifier '${enclaveIdentifier}', service UUIDs '${serviceUuids}', with should follow value '${shouldFollowLogs}' and log line filter '${logLineFilter}'. Error:\n${error}`));
        }

        const streamServiceLogsResult = await this.client.getServiceLogs(getServiceLogsArgs);
        if(streamServiceLogsResult.isErr()){
            return err(streamServiceLogsResult.error)
        }

        const serviceLogsReadable: Readable = streamServiceLogsResult.value;

        return ok(serviceLogsReadable)
    }

    public async getExistingAndHistoricalEnclaveIdentifiers(): Promise<Result<EnclaveIdentifiers, Error>> {
        const getExistingAndHistoricalEnclaveIdentifiersResponseResult = await this.client.getExistingAndHistoricalEnclaveIdentifiers();
        if (getExistingAndHistoricalEnclaveIdentifiersResponseResult.isErr()) {
            return err(getExistingAndHistoricalEnclaveIdentifiersResponseResult.error);
        }

        const getExistingAndHistoricalEnclaveIdentifiersValue = getExistingAndHistoricalEnclaveIdentifiersResponseResult.value

        return ok(new EnclaveIdentifiers(getExistingAndHistoricalEnclaveIdentifiersValue.getAllidentifiersList()))
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    private async newEnclaveContextFromEnclaveInfo(enclaveInfo: EnclaveInfo): Promise<Result<EnclaveContext, Error>> {
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

        let newEnclaveContextResult: Result<EnclaveContext, Error>
        newEnclaveContextResult = await EnclaveContext.newGrpcNodeEnclaveContext(
            LOCAL_HOSTNAME,
            apiContainerHostMachineInfo.getGrpcPortOnHostMachine(),
            enclaveInfo.getEnclaveUuid(),
            enclaveInfo.getName(),
        )
        if(newEnclaveContextResult.isErr()){
            return err(newEnclaveContextResult.error)
        }

        const newEnclaveContext = newEnclaveContextResult.value
        return ok(newEnclaveContext);
    }

    private static async validateEngineApiVersion(genericKurtosisContext: GenericEngineClient): Promise<Result<null, Error>>{
        const getEngineInfoResult = await genericKurtosisContext.getEngineInfo()
        if(getEngineInfoResult.isErr()){
            return err(getEngineInfoResult.error)
        }
        const engineInfoResponse: GetEngineInfoResponse = getEngineInfoResult.value;

        const runningEngineVersionStr: string = engineInfoResponse.getEngineVersion()

        const runningEngineSemver: semver.SemVer | null = semver.parse(runningEngineVersionStr)
        if (runningEngineSemver === null){
            log.warn(`We expected the running engine version to match format X.Y.Z, but instead got ${runningEngineVersionStr}; this means that we can't verify the API library and engine versions match so you may encounter runtime errors`)
        }

        const libraryEngineSemver: semver.SemVer | null = semver.parse(KURTOSIS_VERSION)
        if (libraryEngineSemver === null){
            log.warn(`We expected the library engine version to match format X.Y.Z, but instead got ${KURTOSIS_VERSION}; this means that we can't verify the API library and engine versions match so you may encounter runtime errors`)
        }

        if(runningEngineSemver && libraryEngineSemver){
            const runningEngineMajorVersion = semver.major(runningEngineSemver)
            const runningEngineMinorVersion = semver.minor(runningEngineSemver)

            const libraryEngineMajorVersion = semver.major(libraryEngineSemver)
            const libraryEngineMinorVersion = semver.minor(libraryEngineSemver)

            const doApiVersionsMatch: boolean = libraryEngineMajorVersion === runningEngineMajorVersion && libraryEngineMinorVersion === runningEngineMinorVersion
            if (!doApiVersionsMatch) {
                return err(new Error(
                    `An API version mismatch was detected between the running engine version '${runningEngineSemver.version}' and the engine version this Kurtosis SDK library expects, '${libraryEngineSemver.version}'. You should:\n` +
                    `  1) upgrade your Kurtosis CLI to latest using the instructions at https://docs.kurtosis.com/upgrade\n` +
                    `  2) use the Kurtosis CLI to restart your engine via 'kurtosis engine restart'\n`	+
                    `  3) upgrade your Kurtosis SDK library using the instructions at https://github.com/kurtosis-tech/kurtosis-sdk\n`,
                ));
            }
        }

        return ok(null)
    }

    async createEnclaveWithArgs(enclaveArgs: CreateEnclaveArgs): Promise<Result<EnclaveContext, Error>> {
        const getEnclaveResponseResult = await this.client.createEnclaveResponse(enclaveArgs)
        if(getEnclaveResponseResult.isErr()){
            return err(getEnclaveResponseResult.error)
        }

        const enclaveResponse: CreateEnclaveResponse = getEnclaveResponseResult.value;
        const enclaveInfo: EnclaveInfo | undefined = enclaveResponse.getEnclaveInfo();
        if (enclaveInfo === undefined) {
            return err(new Error("An error occurred creating enclave with name " + enclaveArgs.getEnclaveName() + " enclaveInfo is undefined; " +
                "this is a bug on this library" ))
        }

        const newEnclaveContextResult: Result<EnclaveContext, Error> = await this.newEnclaveContextFromEnclaveInfo(enclaveInfo);
        if (newEnclaveContextResult.isErr()) {
            return err(new Error(`An error occurred creating an enclave context from a newly-created enclave; this should never happen`))
        }

        const enclaveContext = newEnclaveContextResult.value
        return ok(enclaveContext);
    }

}
