/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import { ok, err, Result } from "neverthrow";
import log from "loglevel";
import * as jspb from "google-protobuf";
import * as path from "path-browserify"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import { 
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
    ExecuteBulkCommandsArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientNode } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { GrpcNodeEnclaveContextBackend } from "./grpc_node_enclave_context_backend";
import { GrpcWebEnclaveContextBackend } from "./grpc_web_enclave_context_backend";
import EnclaveContextBackend from "./enclave_context_backend";
import { ModuleContext, ModuleID } from "../modules/module_context";
import { newExecuteBulkCommandsArgs, 
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
import { ContainerConfig, FilesArtifactID } from "../services/container_config";
import { ServiceID } from "../services/service";
import { SharedPath } from "../services/shared_path";
import { ServiceContext } from "../services/service_context";
import { PortProtocol, PortSpec } from "../services/port_spec";
import { PartitionConnection } from "./partition_connection";

export type EnclaveID = string;

export type PartitionID = string;

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)
const DEFAULT_PARTITION_ID: PartitionID = "";

// The path on the user service container where the enclave data dir will be bind-mounted
const SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT: string = "/kurtosis-enclave-data";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class EnclaveContext {

    private readonly backend: EnclaveContextBackend

    // The location on the filesystem where this code is running where the enclave data dir exists
    private readonly enclaveDataDirpath: string;

    constructor(client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode, enclaveId: EnclaveID, enclaveDataDirpath: string){
        if(client instanceof ApiContainerServiceClientWeb){
            this.backend = new GrpcWebEnclaveContextBackend(client, enclaveId)
        }else{
            this.backend = new GrpcNodeEnclaveContextBackend(client, enclaveId)
        }
        this.enclaveDataDirpath = enclaveDataDirpath;
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
        const moduleContext:ModuleContext = new ModuleContext(this.backend.getClient(), moduleId);
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

        const getModuleInfotResult = await this.backend.getModuleInfo(getModuleInfoArgs)
        if(getModuleInfotResult.isErr()){
            return err(getModuleInfotResult.error)
        }
        const moduleContext: ModuleContext = new ModuleContext(this.backend.getClient(), moduleId);

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
                this.backend.getClient(),
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
            this.backend.getClient(),
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
            bodyText: string): Promise<Result<null, Error>> {
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
    public async executeBulkCommands(bulkCommandsJson: string): Promise<Result<null, Error>> {
        const executeBulkCommandsArgs: ExecuteBulkCommandsArgs = newExecuteBulkCommandsArgs(bulkCommandsJson);

        const executeBulkCommandsResult = await this.backend.executeBulkCommands(executeBulkCommandsArgs)
        if(executeBulkCommandsResult.isErr()){
            return err(executeBulkCommandsResult.error)
        }

        const result = executeBulkCommandsResult.value

        return ok(result)
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

        const absFilepathOnThisContainer = path.join(this.enclaveDataDirpath, relativeServiceDirpath);
        const absFilepathOnServiceContainer = path.join(SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT, relativeServiceDirpath);

        const sharedDirectory = new SharedPath(absFilepathOnThisContainer, absFilepathOnServiceContainer);

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
