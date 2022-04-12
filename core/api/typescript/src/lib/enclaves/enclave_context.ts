/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import { ok, err, Result } from "neverthrow";
import log from "loglevel";
import { isNode as  isExecutionEnvNode} from "browser-or-node";
import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import type {
    PartitionConnectionInfo,
    PartitionServices,
    Port,
    RemoveServiceArgs,
    RepartitionArgs,
    StartServiceArgs,
    PartitionConnections,
    LoadModuleArgs,
    UnloadModuleArgs,
    GetModuleInfoArgs,
    RegisterFilesArtifactsArgs,
    RegisterServiceArgs,
    GetServiceInfoArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { GrpcNodeApiContainerClient } from "./grpc_node_api_container_client";
import { GrpcWebApiContainerClient } from "./grpc_web_api_container_client";
import type { GenericApiContainerClient } from "./generic_api_container_client";
import { ModuleContext, ModuleID } from "../modules/module_context";
import { 
    newGetModuleInfoArgs,
    newGetServiceInfoArgs,
    newLoadModuleArgs,
    newPartitionConnections,
    newPartitionServices,
    newPort,
    newRegisterFilesArtifactsArgs,
    newRegisterServiceArgs,
    newRemoveServiceArgs,
    newRepartitionArgs,
    newStartServiceArgs,
    newUnloadModuleArgs,
    newWaitForHttpGetEndpointAvailabilityArgs,
    newWaitForHttpPostEndpointAvailabilityArgs
} from "../constructor_calls";
import type { ContainerConfig, FilesArtifactID } from "../services/container_config";
import type { ServiceID } from "../services/service";
import { SharedPath } from "../services/shared_path";
import { ServiceContext } from "../services/service_context";
import { PortProtocol, PortSpec } from "../services/port_spec";
import type { GenericPathJoiner } from "./generic_path_joiner";
import type { PartitionConnection } from "./partition_connection";

export type EnclaveID = string;
export type PartitionID = string;

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)
const DEFAULT_PARTITION_ID: PartitionID = "";

// The path on the user service container where the enclave data dir will be bind-mounted
const SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT: string = "/kurtosis-enclave-data";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class EnclaveContext {

    private readonly backend: GenericApiContainerClient
    private readonly pathJoiner: GenericPathJoiner
    // The location on the filesystem where this code is running where the enclave data dir exists
    private readonly enclaveDataDirpath: string;

    private constructor(backend: GenericApiContainerClient, enclaveDataDirpath: string, pathJoiner: GenericPathJoiner){
        this.backend = backend;
        this.enclaveDataDirpath = enclaveDataDirpath;
        this.pathJoiner = pathJoiner;
    }

    public static async newGrpcWebEnclaveContext(
            ipAddress: string,
            apiContainerGrpcProxyPortNum: number,
            enclaveId: string,
            enclaveDataDirpath: string
        ): Promise<Result<EnclaveContext, Error>> {

        if(isExecutionEnvNode){
            return err(new Error("It seems you're trying to create Enclave Context from Node environment. Please consider the 'newGrpcNodeEnclaveContext()' method instead."))
        }

        let genericApiContainerClient: GenericApiContainerClient
        let pathJoiner: GenericPathJoiner
        try {
            pathJoiner = await import("path-browserify")
            const apiContainerServiceWeb = await import("../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb")

            const apiContainerGrpcProxyUrl: string = `${ipAddress}:${apiContainerGrpcProxyPortNum}`
            const apiContainerClient = new apiContainerServiceWeb.ApiContainerServiceClient(apiContainerGrpcProxyUrl);
            genericApiContainerClient = new GrpcWebApiContainerClient(apiContainerClient, enclaveId)
        }catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that wasn't an error: " + error
            ));
        }
        
        const enclaveContext = new EnclaveContext(genericApiContainerClient, enclaveDataDirpath, pathJoiner);
        return ok(enclaveContext)
    }

    public static async newGrpcNodeEnclaveContext(
            ipAddress: string,
            apiContainerGrpcPortNum: number,
            enclaveId: string,
            enclaveDataDirpath: string
        ): Promise<Result<EnclaveContext, Error>> {

        if(!isExecutionEnvNode){
            return err(new Error("It seems you're trying to create Enclave Context from Web environment. Please consider the 'newGrpcWebEnclaveContext()' method instead."))
        }

        let genericApiContainerClient: GenericApiContainerClient
        let pathJoiner: GenericPathJoiner
        try {
            pathJoiner = await import( /* webpackIgnore: true */ "path")
            const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")
            const apiContainerServiceNode = await import( /* webpackIgnore: true */ "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb")

            const apiContainerGrpcUrl: string = `${ipAddress}:${apiContainerGrpcPortNum}`
            const apiContainerClient = new apiContainerServiceNode.ApiContainerServiceClient(apiContainerGrpcUrl, grpc_node.credentials.createInsecure());
            genericApiContainerClient = new GrpcNodeApiContainerClient(apiContainerClient, enclaveId)

        }catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that wasn't an error: " + error
            ));
        }

        const enclaveContext = new EnclaveContext(genericApiContainerClient, enclaveDataDirpath, pathJoiner);
        return ok(enclaveContext)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getEnclaveId(): EnclaveID {
        return this.backend.getEnclaveId();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async loadModule(moduleId: ModuleID, image: string, serializedParams: string): Promise<Result<ModuleContext, Error>> {
        const loadModuleArgs: LoadModuleArgs = newLoadModuleArgs(moduleId, image, serializedParams);

        const loadModuleResult = await this.backend.loadModule(loadModuleArgs)
        if(loadModuleResult.isErr()){
            return err(loadModuleResult.error)
        }
        
        const moduleContext:ModuleContext = new ModuleContext(this.backend, moduleId);
        return ok(moduleContext)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async unloadModule(moduleId: ModuleID): Promise<Result<null,Error>> {
        const unloadModuleArgs: UnloadModuleArgs = newUnloadModuleArgs(moduleId);

        const unloadModuleResult = await this.backend.unloadModule(unloadModuleArgs)
        if(unloadModuleResult.isErr()){
            return err(unloadModuleResult.error)
        }

        const result = unloadModuleResult.value
        return ok(result)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>> {
        const getModuleInfoArgs: GetModuleInfoArgs = newGetModuleInfoArgs(moduleId);

        const getModuleInfoResult = await this.backend.getModuleInfo(getModuleInfoArgs)
        if(getModuleInfoResult.isErr()){
            return err(getModuleInfoResult.error)
        }

        const moduleContext: ModuleContext = new ModuleContext(this.backend, moduleId);
        return ok(moduleContext)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async registerFilesArtifacts(filesArtifactUrls: Map<FilesArtifactID, string>): Promise<Result<null,Error>> {
        const filesArtifactIdStrsToUrls: Map<string, string> = new Map();
        for (const [artifactId, url] of filesArtifactUrls.entries()) {
            filesArtifactIdStrsToUrls.set(String(artifactId), url);
        }
        const registerFilesArtifactsArgs: RegisterFilesArtifactsArgs = newRegisterFilesArtifactsArgs(filesArtifactIdStrsToUrls);

        const registerFilesArtifactsResult = await this.backend.registerFilesArtifacts(registerFilesArtifactsArgs)

        if(registerFilesArtifactsResult.isErr()){
            return err(registerFilesArtifactsResult.error)
        }
        const result = registerFilesArtifactsResult.value
        return ok(result)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async addService(
            serviceId: ServiceID,
            containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error>
        ): Promise<Result<ServiceContext, Error>> {

        const resultAddServiceToPartition: Result<ServiceContext, Error> = await this.addServiceToPartition(
            serviceId,
            DEFAULT_PARTITION_ID,
            containerConfigSupplier,
        );
        if (resultAddServiceToPartition.isErr()) {
            return err(resultAddServiceToPartition.error);
        }

        return ok(resultAddServiceToPartition.value);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async addServiceToPartition(
            serviceId: ServiceID,
            partitionId: PartitionID,
            containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error>
        ): Promise<Result<ServiceContext, Error>> {

        log.trace("Registering new service ID with Kurtosis API...");

        const registerServiceArgs: RegisterServiceArgs = newRegisterServiceArgs(serviceId, partitionId);

        const registerServiceResponseResult = await this.backend.registerService(registerServiceArgs)
        if(registerServiceResponseResult.isErr()){
            return err(registerServiceResponseResult.error)
        }

        const registerServiceResponse = registerServiceResponseResult.value

        log.trace("New service successfully registered with Kurtosis API");

        const privateIpAddr: string = registerServiceResponse.getPrivateIpAddr();
        const relativeServiceDirpath: string = registerServiceResponse.getRelativeServiceDirpath();

        const sharedDirectory = this.getSharedDirectory(relativeServiceDirpath)

        log.trace("Generating container config object using the container config supplier...")
        const containerConfigSupplierResult: Result<ContainerConfig, Error> = containerConfigSupplier(privateIpAddr, sharedDirectory);
        if (containerConfigSupplierResult.isErr()){
            return err(containerConfigSupplierResult.error);
        }
        const containerConfig: ContainerConfig = containerConfigSupplierResult.value;
        log.trace("Container config object successfully generated")

        log.trace("Creating files artifact ID str -> mount dirpaths map...");
        const artifactIdStrToMountDirpath: Map<string, string> = new Map();
        for (const [filesArtifactId, mountDirpath] of containerConfig.filesArtifactMountpoints.entries()) {

            artifactIdStrToMountDirpath.set(String(filesArtifactId), mountDirpath);
        }
        log.trace("Successfully created files artifact ID str -> mount dirpaths map");

        log.trace("Starting new service with Kurtosis API...");
        const privatePorts = containerConfig.usedPorts;
        const privatePortsForApi: Map<string, Port> = new Map();
        for (const [portId, portSpec] of privatePorts.entries()) {
            const portSpecForApi: Port = newPort(
                portSpec.number,
                portSpec.protocol,
            )
            privatePortsForApi.set(portId, portSpecForApi);
        }
        const startServiceArgs: StartServiceArgs = newStartServiceArgs(
            serviceId,
            containerConfig.image,
            privatePortsForApi,
            containerConfig.entrypointOverrideArgs,
            containerConfig.cmdOverrideArgs,
            containerConfig.environmentVariableOverrides,
            SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT,
            artifactIdStrToMountDirpath);

        const startServiceResponseResult = await this.backend.startService(startServiceArgs)
        if(startServiceResponseResult.isErr()){
            return err(startServiceResponseResult.error)
        }

        const startServiceResponse = startServiceResponseResult.value

        log.trace("Successfully started service with Kurtosis API");

        const serviceCtxPublicPorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            startServiceResponse.getPublicPortsMap(),
        );

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend,
            serviceId,
            sharedDirectory,
            privateIpAddr,
            privatePorts,
            startServiceResponse.getPublicIpAddr(),
            serviceCtxPublicPorts,
        );

        return ok(serviceContext)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getServiceContext(serviceId: ServiceID): Promise<Result<ServiceContext, Error>> {
        const getServiceInfoArgs: GetServiceInfoArgs = newGetServiceInfoArgs(serviceId);

        const getServiceInfoResult = await this.backend.getServiceInfo(getServiceInfoArgs)
        if(getServiceInfoResult.isErr()){
            return err(getServiceInfoResult.error)
        }

        const serviceInfo = getServiceInfoResult.value
        if (serviceInfo.getPrivateIpAddr() === "") {
            return err(new Error(
                    "Kurtosis API reported an empty private IP address for service " + serviceId +  " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }
        if (serviceInfo.getPublicIpAddr() === "") {
            return err(new Error(
                    "Kurtosis API reported an empty public IP address for service " + serviceId +  " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const relativeServiceDirpath: string = serviceInfo.getRelativeServiceDirpath();
        if (relativeServiceDirpath === "") {
            return err(new Error(
                    "Kurtosis API reported an empty relative service directory path for service " + serviceId + " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const enclaveDataDirMountDirpathOnSvcContainer: string = serviceInfo.getEnclaveDataDirMountDirpath();
        if (enclaveDataDirMountDirpathOnSvcContainer === "") {
            return err(new Error(
                    "Kurtosis API reported an empty enclave data dir mount dirpath for service " + serviceId + " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const sharedDirectory: SharedPath = this.getSharedDirectory(relativeServiceDirpath)

        const serviceCtxPrivatePorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getPrivatePortsMap(),
        );
        const serviceCtxPublicPorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getPublicPortsMap(),
        );

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend,
            serviceId,
            sharedDirectory,
            serviceInfo.getPrivateIpAddr(),
            serviceCtxPrivatePorts,
            serviceInfo.getPublicIpAddr(),
            serviceCtxPublicPorts,
        );

        return ok(serviceContext);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async removeService(serviceId: ServiceID, containerStopTimeoutSeconds: number): Promise<Result<null, Error>> {
        log.debug("Removing service '" + serviceId + "'...");
        // NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
        //  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
        // Independent of adding/removing them from the enclave
        const removeServiceArgs: RemoveServiceArgs = newRemoveServiceArgs(serviceId, containerStopTimeoutSeconds);

        const removeServiceResult = await this.backend.removeService(removeServiceArgs)
        if(removeServiceResult.isErr()){
            return err(removeServiceResult.error)
        }

        log.debug("Successfully removed service ID " + serviceId);

        return ok(null)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async repartitionNetwork(
            partitionServices: Map<PartitionID, Set<ServiceID>>,
            partitionConnections: Map<PartitionID, Map<PartitionID, PartitionConnection>>,
            defaultConnection: PartitionConnection
        ): Promise<Result<null, Error>> {

        if (partitionServices === null) {
            return err(new Error("Partition services map cannot be null"));
        }
        if (defaultConnection === null) {
            return err(new Error("Default connection cannot be null"));
        }

        // Cover for lazy/confused users
        if (partitionConnections === null) {
            partitionConnections = new Map();
        }

        const reqPartitionServices: Map<string, PartitionServices> = new Map();
        for (const [partitionId, serviceIdSet] of partitionServices.entries()) {
            const partitionIdStr: string = String(partitionId);
            reqPartitionServices.set(partitionIdStr, newPartitionServices(serviceIdSet));
        }

        const reqPartitionConns: Map<string, PartitionConnections> = new Map();
        for (const [partitionAId, partitionAConnsMap] of partitionConnections.entries()) {
            const partitionAConnsStrMap: Map<string, PartitionConnectionInfo> = new Map();

            for (const [partitionBId, conn] of partitionAConnsMap.entries()) {
                const partitionBIdStr: string = String(partitionBId);
                partitionAConnsStrMap.set(partitionBIdStr, conn.getPartitionConnectionInfo());
            }

            const partitionAConns: PartitionConnections = newPartitionConnections(partitionAConnsStrMap);
            const partitionAIdStr: string = String(partitionAId);
            reqPartitionConns.set(partitionAIdStr, partitionAConns);
        }

        const reqDefaultConnection = defaultConnection.getPartitionConnectionInfo()

        const repartitionArgs: RepartitionArgs = newRepartitionArgs(reqPartitionServices, reqPartitionConns, reqDefaultConnection);

        const repartitionNetworkResult = await this.backend.repartitionNetwork(repartitionArgs)
        if(repartitionNetworkResult.isErr()){
            return err(repartitionNetworkResult.error)
        }

        return ok(null)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async waitForHttpGetEndpointAvailability(
            serviceId: ServiceID,
            port: number,
            path: string,
            initialDelayMilliseconds: number,
            retries: number,
            retriesDelayMilliseconds: number,
            bodyText: string
        ): Promise<Result<null, Error>> {

        const availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs = newWaitForHttpGetEndpointAvailabilityArgs(
            serviceId,
            port,
            path,
            initialDelayMilliseconds,
            retries,
            retriesDelayMilliseconds,
            bodyText
        );

        const waitForHttpGetEndpointAvailabilityResult = await this.backend.waitForHttpGetEndpointAvailability(availabilityArgs)
        if(waitForHttpGetEndpointAvailabilityResult.isErr()){
            return err(waitForHttpGetEndpointAvailabilityResult.error)
        }

        const result = waitForHttpGetEndpointAvailabilityResult.value
        return ok(result)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async waitForHttpPostEndpointAvailability(
            serviceId: ServiceID,
            port: number,
            path: string,
            requestBody: string,
            initialDelayMilliseconds: number,
            retries: number,
            retriesDelayMilliseconds: number,
            bodyText: string
        ): Promise<Result<null, Error>> {
        const availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs = newWaitForHttpPostEndpointAvailabilityArgs(
            serviceId,
            port,
            path,
            requestBody,
            initialDelayMilliseconds,
            retries,
            retriesDelayMilliseconds,
            bodyText);

        return this.backend.waitForHttpPostEndpointAvailability(availabilityArgs)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getServices(): Promise<Result<Set<ServiceID>, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getServicesResponseResult = await this.backend.getServices(emptyArg)
        if(getServicesResponseResult.isErr()){
            return err(getServicesResponseResult.error)
        }

        const getServicesResponse = getServicesResponseResult.value

        const serviceIDs: Set<ServiceID> = new Set<ServiceID>()

        getServicesResponse.getServiceIdsMap().forEach((value: boolean, key: string) => {
            serviceIDs.add(key)
        });

        return ok(serviceIDs)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModules(): Promise<Result<Set<ModuleID>, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getModulesResponseResult = await this.backend.getModules(emptyArg)
        if(getModulesResponseResult.isErr()){
            return err(getModulesResponseResult.error)
        }

        const modulesResponse = getModulesResponseResult.value

        const moduleIds: Set<ModuleID> = new Set<ModuleID>()

        modulesResponse.getModuleIdsMap().forEach((value: boolean, key: string) => {
            moduleIds.add(key)
        })

        return ok(moduleIds)
    }


    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    private getSharedDirectory(relativeServiceDirpath: string): SharedPath {

        const absFilepathOnThisContainer = this.pathJoiner.join(this.enclaveDataDirpath, relativeServiceDirpath);
        const absFilepathOnServiceContainer = this.pathJoiner.join(SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT, relativeServiceDirpath);

        const sharedDirectory = new SharedPath(absFilepathOnThisContainer, absFilepathOnServiceContainer, this.pathJoiner);
        return sharedDirectory;
    }

    private static convertApiPortsToServiceContextPorts(apiPorts: jspb.Map<string, Port>): Map<string, PortSpec> {
        const result: Map<string, PortSpec> = new Map();
        for (const [portId, apiPortSpec] of apiPorts.entries()) {
            const portProtocol: PortProtocol = apiPortSpec.getProtocol();
            const portNum: number = apiPortSpec.getNumber();
            result.set(portId, new PortSpec(portNum, portProtocol))
        }
        return result;
    }
}
