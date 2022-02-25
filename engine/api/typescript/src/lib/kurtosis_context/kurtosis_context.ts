import log from "loglevel"
import * as semver from "semver"
import * as jspb from "google-protobuf";
import {err, ok, Result} from "neverthrow";
import { isNode as isExecutionEnvNode} from "browser-or-node";
import { EnclaveContext, EnclaveID } from "kurtosis-core-api-lib";
import { GenericEngineClient } from "./generic_engine_client";
import { KURTOSIS_ENGINE_VERSION } from "../../kurtosis_engine_version/kurtosis_engine_version";
import { GrpcWebEngineClient } from "./grpc_web_engine_client";
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
    StopEnclaveArgs
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import { newCleanArgs, newCreateEnclaveArgs, newDestroyEnclaveArgs, newStopEnclaveArgs } from "../constructor_calls";

//It seems that gRPC web vs gRPC node read/access differently localhost. 
//'0.0.0.0' works in Node, but doesn't work in Web and viceversa.
const LOCAL_NODE_HOST_IP_ADDRESS_STR: string = "0.0.0.0";
const LOCAL_HOST: string = "localhost";

const SHOULD_PUBLISH_ALL_PORTS: boolean = true;

const API_CONTAINER_LOG_LEVEL: string = "debug";

export const DEFAULT_GRPC_PROXY_ENGINE_SERVER_PORT_NUM: number = 9711;
export const DEFAULT_GRPC_ENGINE_SERVER_PORT_NUM: number = 9710;

// Blank tells the engine server to use the default
const DEFAULT_API_CONTAINER_VERSION_TAG = "";

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
export class KurtosisContext {
    private readonly backend: GenericEngineClient

    constructor(backend: GenericEngineClient){
        this.backend = backend;
    }

    // Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
    public static async newKurtosisContextFromLocalEngine():Promise<Result<KurtosisContext, Error>>  {
        let genericKurtosisContextBackend: GenericEngineClient
        try {
            if(isExecutionEnvNode){
                const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")
                const engineServiceNode = await import( /* webpackIgnore: true */ "../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb")

                const kurtosisEngineSocketStr: string = `${LOCAL_NODE_HOST_IP_ADDRESS_STR}:${DEFAULT_GRPC_ENGINE_SERVER_PORT_NUM}`
                const engineServiceClientNode = new engineServiceNode.EngineServiceClient(kurtosisEngineSocketStr, grpc_node.credentials.createInsecure())
                genericKurtosisContextBackend = new GrpcNodeEngineClient(engineServiceClientNode)
            }else {
                const engineServiceWeb = await import("../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb")

                const kurtosisEngineSocketStr: string = `http://${LOCAL_HOST}:${DEFAULT_GRPC_PROXY_ENGINE_SERVER_PORT_NUM}`
                const engineServiceClientWeb = new engineServiceWeb.EngineServiceClient(kurtosisEngineSocketStr)
                genericKurtosisContextBackend = new GrpcWebEngineClient(engineServiceClientWeb)
            }
        } catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the engine client that wasn't an error: " + error
            ));
        }

        const getEngineInfoResult = await this.getEngineInfo(genericKurtosisContextBackend)
        if(getEngineInfoResult.isErr()){
            return err(getEngineInfoResult.error)
        }

        const kurtosisContext = new KurtosisContext(genericKurtosisContextBackend)
        return ok(kurtosisContext)
    }

    public async createEnclave(enclaveId: string, isPartitioningEnabled: boolean): Promise<Result<EnclaveContext, Error>> {
        const enclaveArgs: CreateEnclaveArgs = newCreateEnclaveArgs(
            enclaveId,
            DEFAULT_API_CONTAINER_VERSION_TAG,
            API_CONTAINER_LOG_LEVEL,
            isPartitioningEnabled,
            SHOULD_PUBLISH_ALL_PORTS,
        );

        const getEnclaveResponseResult = await this.backend.createEnclaveResponse(enclaveArgs)
        if(getEnclaveResponseResult.isErr()){
            return err(getEnclaveResponseResult.error)
        }

        const enclaveResponse: CreateEnclaveResponse = getEnclaveResponseResult.value;
        const enclaveInfo: EnclaveInfo | undefined = enclaveResponse.getEnclaveInfo();
        if (enclaveInfo === undefined) {
            return err(new Error("An error occurred creating enclave with ID " + enclaveId + " enclaveInfo is undefined; " +
                "this is a bug on this library" ))
        }

        const newEnclaveContextResult: Result<EnclaveContext, Error> = await this.newEnclaveContextFromEnclaveInfo(enclaveInfo);
        if (newEnclaveContextResult.isErr()) {
            return err(new Error(`An error occurred creating an enclave context from a newly-created enclave; this should never happen`))
        }

        const enclaveContext = newEnclaveContextResult.value
        return ok(enclaveContext);
    }

    public async getEnclaveContext(enclaveId: EnclaveID): Promise<Result<EnclaveContext, Error>> {
        const getEnclavesResponseResult = await this.backend.getEnclavesResponse();
        if (getEnclavesResponseResult.isErr()) {
            return err(getEnclavesResponseResult.error);
        }
        const getEnclavesResponse: GetEnclavesResponse = getEnclavesResponseResult.value;

        const allEnclaveInfo: jspb.Map<string, EnclaveInfo> = getEnclavesResponse.getEnclaveInfoMap()
        const maybeEnclaveInfo: EnclaveInfo | undefined = allEnclaveInfo.get(enclaveId);
        if (maybeEnclaveInfo === undefined) {
            return err(new Error(`No enclave with ID '${enclaveId}' found`))
        }

        const newEnclaveContextResult: Result<EnclaveContext, Error> = await this.newEnclaveContextFromEnclaveInfo(maybeEnclaveInfo);
        if (newEnclaveContextResult.isErr()) {
            return err(newEnclaveContextResult.error);
        }

        return ok(newEnclaveContextResult.value);
    }

    public async getEnclaves(): Promise<Result<Set<EnclaveID>, Error>>{
        const getEnclavesResponseResult = await this.backend.getEnclavesResponse();
        if (getEnclavesResponseResult.isErr()) {
            return err(getEnclavesResponseResult.error);
        }

        const getEnclavesResponse: GetEnclavesResponse = getEnclavesResponseResult.value;
        const enclaves: Set<EnclaveID> = new Set();
        for (let enclaveId of getEnclavesResponse.getEnclaveInfoMap().keys()) {
            enclaves.add(enclaveId);
        }

        return ok(enclaves);
    }

    public async stopEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>>{
        const stopEnclaveArgs: StopEnclaveArgs = newStopEnclaveArgs(enclaveId)
        const stopEnclaveResult = await this.backend.stopEnclave(stopEnclaveArgs)
        if(stopEnclaveResult.isErr()){
            return err(stopEnclaveResult.error)
        }

        return ok(null)
    }

    public async destroyEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>>{
        const destroyEnclaveArgs: DestroyEnclaveArgs = newDestroyEnclaveArgs(enclaveId);
        const destroyEnclaveResult = await this.backend.destroyEnclave(destroyEnclaveArgs)
        if(destroyEnclaveResult.isErr()){
            return err(destroyEnclaveResult.error)
        }

        return ok(null)
    }

    public async clean(shouldCleanAll : boolean): Promise<Result<Set<string>, Error>>{
        const cleanArgs: CleanArgs = newCleanArgs(shouldCleanAll);
        const cleanResponseResult = await this.backend.clean(cleanArgs)
        if(cleanResponseResult.isErr()){
            return err(cleanResponseResult.error)
        }

        const cleanResponse: CleanResponse = cleanResponseResult.value
        const result: Set<string> = new Set();
        for (let enclaveID of cleanResponse.getRemovedEnclaveIdsMap().keys()) {
            result.add(enclaveID);
        }

        return ok(result)
    }


    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    private static async getEngineInfo(genericKurtosisContext: GenericEngineClient): Promise<Result<null, Error>>{
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

        const libraryEngineSemver: semver.SemVer | null = semver.parse(KURTOSIS_ENGINE_VERSION)
        if (libraryEngineSemver === null){
            log.warn(`We expected the library engine version to match format X.Y.Z, but instead got ${KURTOSIS_ENGINE_VERSION}; this means that we can't verify the API library and engine versions match so you may encounter runtime errors`)
        }

        if(runningEngineSemver && libraryEngineSemver){
            const runningEngineMajorVersion = semver.major(runningEngineSemver)
            const runningEngineMinorVersion = semver.minor(runningEngineSemver)

            const libraryEngineMajorVersion = semver.major(libraryEngineSemver)
            const libraryEngineMinorVersion = semver.minor(libraryEngineSemver)

            const doApiVersionsMatch: boolean = libraryEngineMajorVersion === runningEngineMajorVersion && libraryEngineMinorVersion === runningEngineMinorVersion
            if (!doApiVersionsMatch) {
                return err(new Error(
                    `An API version mismatch was detected between the running engine version ${runningEngineSemver.version} and the engine version the library expects, ${libraryEngineSemver.version}; you should use the version of this library that corresponds to the running engine version`
                ));
            }
        }

        return ok(null)
    }

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
        try{
            if(isExecutionEnvNode){
                newEnclaveContextResult = await EnclaveContext.newGrpcNodeEnclaveContext(
                    LOCAL_NODE_HOST_IP_ADDRESS_STR,
                    apiContainerHostMachineInfo.getGrpcPortOnHostMachine(),
                    enclaveInfo.getEnclaveId(),
                    enclaveInfo.getEnclaveDataDirpathOnHostMachine(),
                    )
            }else{
                newEnclaveContextResult = await EnclaveContext.newGrpcWebEnclaveContext(
                    LOCAL_HOST,
                    apiContainerHostMachineInfo.getGrpcProxyPortOnHostMachine(),
                    enclaveInfo.getEnclaveId(),
                    enclaveInfo.getEnclaveDataDirpathOnHostMachine(),
                )
            }
        }catch(error){
            if(error instanceof Error){
                return err(error)
            }else{
                return err(new Error(
                    "An unknown exception value was thrown during creation of the Enclave context that wasn't an error: " + error
                ))
            }
        }
        if(newEnclaveContextResult.isErr()){
            return err(newEnclaveContextResult.error)
        }

        const newEnclaveContext = newEnclaveContextResult.value
        return ok(newEnclaveContext);
    }
}