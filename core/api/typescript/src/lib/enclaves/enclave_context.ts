/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import { ok, err, Result } from "neverthrow";
import log from "loglevel";
import * as jspb from "google-protobuf";
import * as path from "path-browserify"
import { 
    ApiContainerServiceClientWeb, 
    ApiContainerServiceClientNode, 
    ModuleID, 
    ModuleContext,
    ServiceID,
    SharedPath,
    ServiceContext,
    ContainerConfig, 
    FilesArtifactID,
    PartitionConnection,
    PartitionConnectionInfo, 
    PartitionServices, 
    Port, 
    RemoveServiceArgs, 
    RepartitionArgs, 
    StartServiceArgs,
    PartitionConnections,
    RegisterServiceResponse,
    StartServiceResponse,
    GetServiceInfoResponse,
    GetModulesResponse,
    GetServicesResponse,
    newPartitionConnections, 
    newPartitionServices, 
    newPort, 
    newRemoveServiceArgs, 
    newRepartitionArgs, 
    newStartServiceArgs,
    PortProtocol, PortSpec
} from "../../index";
import { GrpcNodeEnclaveContextBackend } from "./enclave_context_backend_node";
import { GrpcWebEnclaveContextBackend } from "./enclave_context_backend_web";

export type EnclaveID = string;

export type PartitionID = string;

const DEFAULT_PARTITION_ID: PartitionID = "";

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)

// The path on the user service container where the enclave data dir will be bind-mounted
const SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT: string = "/kurtosis-enclave-data";

export interface EnclaveContextBackend {
    client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode
    getEnclaveId(): EnclaveID
    loadModule(moduleId: ModuleID, image: string, serializedParams: string): Promise<Result<ModuleContext, Error>>
    unloadModule(moduleId: ModuleID): Promise<Result<null,Error>>
    getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>>
    registerFilesArtifacts(filesArtifactUrls: Map<FilesArtifactID, string>): Promise<Result<null,Error>>
    registerService( serviceId: ServiceID, partitionId: PartitionID): Promise<Result<RegisterServiceResponse, Error>>
    startService(startServiceArgs: StartServiceArgs): Promise<Result<StartServiceResponse, Error>>
    getServiceInfo(serviceId: ServiceID): Promise<Result<GetServiceInfoResponse, Error>>
    removeService(args: RemoveServiceArgs): Promise<Result<null, Error>>
    repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>>
    waitForHttpGetEndpointAvailability(
        serviceId: ServiceID,
        port: number, 
        path: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(
        serviceId: ServiceID,
        port: number, 
        path: string,
        requestBody: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): Promise<Result<null, Error>>
    executeBulkCommands(bulkCommandsJson: string): Promise<Result<null, Error>>
    getServices(): Promise<Result<GetServicesResponse, Error>>
    getModules(): Promise<Result<GetModulesResponse, Error>>
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class EnclaveContext {

    private backend: EnclaveContextBackend

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
       return this.backend.loadModule(moduleId, image, serializedParams)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async unloadModule(moduleId: ModuleID): Promise<Result<null,Error>> {
       return this.backend.unloadModule(moduleId)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>> {
       return this.backend.getModuleContext(moduleId)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async registerFilesArtifacts(filesArtifactUrls: Map<FilesArtifactID, string>): Promise<Result<null,Error>> {
      return this.backend.registerFilesArtifacts(filesArtifactUrls)
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

            const registerServiceResponseResult = await this.backend.registerService(serviceId, partitionId)

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
    
            const serviceCtxPublicPorts: Map<string, PortSpec> = this.convertApiPortsToServiceContextPorts(
                startServiceResponse.getPublicPortsMap(),
            );
    
            const serviceContext: ServiceContext = new ServiceContext(
                this.backend.client,
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

        const getServiceInfoResult = await this.backend.getServiceInfo(serviceId)

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

        const serviceCtxPrivatePorts: Map<string, PortSpec> = this.convertApiPortsToServiceContextPorts(
            serviceInfo.getPrivatePortsMap(),
        );
        const serviceCtxPublicPorts: Map<string, PortSpec> = this.convertApiPortsToServiceContextPorts(
            serviceInfo.getPublicPortsMap(),
        );

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend.client,
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
        const RemoveServiceArgs: RemoveServiceArgs = newRemoveServiceArgs(serviceId, containerStopTimeoutSeconds);

        const removeServiceResult = await this.backend.removeService(RemoveServiceArgs)

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
                return err(new Error("Partition services map cannot be nil"));
            }
            if (defaultConnection === null) {
                return err(new Error("Default connection cannot be nil"));
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
            bodyText: string): Promise<Result<null, Error>> {
            
        return this.backend,this.waitForHttpGetEndpointAvailability(serviceId, port, path, initialDelayMilliseconds, retries, retriesDelayMilliseconds, bodyText)
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
                
        return this.backend,this.waitForHttpPostEndpointAvailability(serviceId, port, path, requestBody,initialDelayMilliseconds, retries, retriesDelayMilliseconds, bodyText)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async executeBulkCommands(bulkCommandsJson: string): Promise<Result<null, Error>> {
       return this.backend.executeBulkCommands(bulkCommandsJson)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getServices(): Promise<Result<Set<ServiceID>, Error>> {
        const getServicesResponseResult = await  this.backend.getServices()

        if(getServicesResponseResult.isErr()){
            return err(getServicesResponseResult.error)
        }

        const servicesResponse = getServicesResponseResult.value

        const serviceIDs: Set<ServiceID> = new Set<ServiceID>()

        servicesResponse.getServiceIdsMap().forEach((value: boolean, key: string) => {
            serviceIDs.add(key)
        });

        return ok(serviceIDs)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModules(): Promise<Result<Set<ModuleID>, Error>> {
        const getModulesResponseResult = await this.backend.getModules()

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

    private convertApiPortsToServiceContextPorts(apiPorts: jspb.Map<string, Port>): Map<string, PortSpec> {
        const result: Map<string, PortSpec> = new Map();
        for (const [portId, apiPortSpec] of apiPorts.entries()) {
            const portProtocol: PortProtocol = apiPortSpec.getProtocol();
            const portNum: number = apiPortSpec.getNumber();
            result.set(portId, new PortSpec(portNum, portProtocol))
        }
        return result;
    }
}
