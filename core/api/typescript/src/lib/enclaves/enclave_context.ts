/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import {ok, err, Result, Err} from "neverthrow";
import log from "loglevel";
import { isNode as  isExecutionEnvNode} from "browser-or-node";
import * as jspb from "google-protobuf";
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
    GetModulesArgs,
    RegisterServiceArgs,
    GetServicesArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { GrpcNodeApiContainerClient } from "./grpc_node_api_container_client";
import { GrpcWebApiContainerClient } from "./grpc_web_api_container_client";
import type { GenericApiContainerClient } from "./generic_api_container_client";
import { ModuleContext, ModuleID } from "../modules/module_context";
import {
    newGetModulesArgs,
    newGetServicesArgs,
    newLoadModuleArgs,
    newPartitionConnections,
    newPartitionServices,
    newPort,
    newRegisterServiceArgs,
    newRemoveServiceArgs,
    newRepartitionArgs,
    newStartServiceArgs,
    newStoreWebFilesArtifactArgs,
    newStoreFilesArtifactFromServiceArgs,
    newUnloadModuleArgs,
    newWaitForHttpGetEndpointAvailabilityArgs,
    newWaitForHttpPostEndpointAvailabilityArgs,
    newUploadFilesArtifactArgs, newPauseServiceArgs, newUnpauseServiceArgs
} from "../constructor_calls";
import type { ContainerConfig, FilesArtifactID } from "../services/container_config";
import type { ServiceID } from "../services/service";
import { ServiceContext } from "../services/service_context";
import { PortProtocol, PortSpec } from "../services/port_spec";
import type { GenericPathJoiner } from "./generic_path_joiner";
import type { PartitionConnection } from "./partition_connection";
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {
    ModuleInfo,
    PauseServiceArgs, ServiceInfo, UnloadModuleResponse,
    UnpauseServiceArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export type EnclaveID = string;
export type PartitionID = string;

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)
const DEFAULT_PARTITION_ID: PartitionID = "";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class EnclaveContext {

    private readonly backend: GenericApiContainerClient
    private readonly pathJoiner: GenericPathJoiner
    private readonly genericTgzArchiver: GenericTgzArchiver

    private constructor(backend: GenericApiContainerClient, pathJoiner: GenericPathJoiner,
                        genericTgzArchiver: GenericTgzArchiver){
        this.backend = backend;
        this.pathJoiner = pathJoiner;
        this.genericTgzArchiver = genericTgzArchiver
    }

    public static async newGrpcWebEnclaveContext(
        ipAddress: string,
        apiContainerGrpcProxyPortNum: number,
        enclaveId: string,
    ): Promise<Result<EnclaveContext, Error>> {

        if(isExecutionEnvNode){
            return err(new Error("It seems you're trying to create Enclave Context from Node environment. Please consider the 'newGrpcNodeEnclaveContext()' method instead."))
        }

        let genericApiContainerClient: GenericApiContainerClient
        let genericTgzArchiver: GenericTgzArchiver
        let pathJoiner: GenericPathJoiner
        try {

            pathJoiner = await import("path-browserify")
            const apiContainerServiceWeb = await import("../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb")

            const apiContainerGrpcProxyUrl: string = `${ipAddress}:${apiContainerGrpcProxyPortNum}`
            const apiContainerClient = new apiContainerServiceWeb.ApiContainerServiceClient(apiContainerGrpcProxyUrl);
            genericApiContainerClient = new GrpcWebApiContainerClient(apiContainerClient, enclaveId)

            const webFileArchiver = await import("./web_tgz_archiver")
            genericTgzArchiver = new webFileArchiver.WebTgzArchiver()
        }catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that wasn't an error: " + error
            ));
        }
        
        const enclaveContext = new EnclaveContext(genericApiContainerClient, pathJoiner, genericTgzArchiver);
        return ok(enclaveContext)
    }

    public static async newGrpcNodeEnclaveContext(
        ipAddress: string,
        apiContainerGrpcPortNum: number,
        enclaveId: string,
    ): Promise<Result<EnclaveContext, Error>> {

        if(!isExecutionEnvNode){
            return err(new Error("It seems you're trying to create Enclave Context from Web environment. Please consider the 'newGrpcWebEnclaveContext()' method instead."))
        }

        let genericApiContainerClient: GenericApiContainerClient
        let genericTgzArchiver: GenericTgzArchiver
        let pathJoiner: GenericPathJoiner
        //TODO Pull things that can't throw an error out of try statement.
        try {
            pathJoiner = await import( /* webpackIgnore: true */ "path")
            const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")
            const apiContainerServiceNode = await import( /* webpackIgnore: true */ "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb")

            const apiContainerGrpcUrl: string = `${ipAddress}:${apiContainerGrpcPortNum}`
            const apiContainerClient = new apiContainerServiceNode.ApiContainerServiceClient(apiContainerGrpcUrl, grpc_node.credentials.createInsecure());
            genericApiContainerClient = new GrpcNodeApiContainerClient(apiContainerClient, enclaveId)

            const nodeTgzArchiver = await import(/* webpackIgnore: true */ "./node_tgz_archiver")
            genericTgzArchiver = new nodeTgzArchiver.NodeTgzArchiver()
        }catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that wasn't an error: " + error
            ));
        }

        const enclaveContext = new EnclaveContext(genericApiContainerClient, pathJoiner, genericTgzArchiver);
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
    public async unloadModule(moduleId: ModuleID): Promise<Result<null ,Error>> {
        const unloadModuleArgs: UnloadModuleArgs = newUnloadModuleArgs(moduleId);

        const unloadModuleResult = await this.backend.unloadModule(unloadModuleArgs)
        if(unloadModuleResult.isErr()){
            return err(unloadModuleResult.error)
        }

        // We discard the module GUID
        return ok(null)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>> {
        const moduleMapForArgs = new Map<string, boolean>()
        moduleMapForArgs.set(moduleId, true)
        const args: GetModulesArgs = newGetModulesArgs(moduleMapForArgs);

        const getModuleInfoResult = await this.backend.getModules(args)
        if(getModuleInfoResult.isErr()){
            return err(getModuleInfoResult.error)
        }
        const resp = getModuleInfoResult.value

        if (!resp.getModuleInfoMap().has(moduleId)) {
            return err(new Error(`Module '${moduleId}' does not exist`))
        }

        const moduleCtx: ModuleContext = new ModuleContext(this.backend, moduleId);
        return ok(moduleCtx)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async addService(
            serviceId: ServiceID,
            containerConfigSupplier: (ipAddr: string) => Result<ContainerConfig, Error>
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
            containerConfigSupplier: (ipAddr: string) => Result<ContainerConfig, Error>
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

        log.trace("Generating container config object using the container config supplier...")
        const containerConfigSupplierResult: Result<ContainerConfig, Error> = containerConfigSupplier(privateIpAddr);
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
            artifactIdStrToMountDirpath,
        );

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
            privateIpAddr,
            privatePorts,
            startServiceResponse.getPublicIpAddr(),
            serviceCtxPublicPorts,
        );

        return ok(serviceContext)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getServiceContext(serviceId: ServiceID): Promise<Result<ServiceContext, Error>> {
        const serviceArgMap = new Map<string, boolean>()
        serviceArgMap.set(serviceId, true)
        const getServiceInfoArgs: GetServicesArgs = newGetServicesArgs(serviceArgMap);

        const getServicesResult = await this.backend.getServices(getServiceInfoArgs)
        if(getServicesResult.isErr()){
            return err(getServicesResult.error)
        }

        const serviceInfo = getServicesResult.value.getServiceInfoMap().get(serviceId)
        if(!serviceInfo) {
            return err(new Error(
                    "Failed to retrieve service information for service " + serviceId
            ))
        }
        if (serviceInfo.getPrivateIpAddr() === "") {
            return err(new Error(
                    "Kurtosis API reported an empty private IP address for service " + serviceId +  " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }
        if (serviceInfo.getMaybePublicIpAddr() === "") {
            return err(new Error(
                    "Kurtosis API reported an empty public IP address for service " + serviceId +  " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const serviceCtxPrivatePorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getPrivatePortsMap(),
        );
        const serviceCtxPublicPorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getMaybePublicPortsMap(),
        );

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend,
            serviceId,
            serviceInfo.getPrivateIpAddr(),
            serviceCtxPrivatePorts,
            serviceInfo.getMaybePublicIpAddr(),
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
        const getAllServicesArgMap: Map<string, boolean> = new Map<string,boolean>()
        const emptyGetServicesArg: GetServicesArgs = newGetServicesArgs(getAllServicesArgMap)

        const getServicesResponseResult = await this.backend.getServices(emptyGetServicesArg)
        if(getServicesResponseResult.isErr()){
            return err(getServicesResponseResult.error)
        }

        const getServicesResponse = getServicesResponseResult.value

        const serviceIDs: Set<ServiceID> = new Set<ServiceID>()

        getServicesResponse.getServiceInfoMap().forEach((value: ServiceInfo, key: string) => {
            serviceIDs.add(key)
        });

        return ok(serviceIDs)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async getModules(): Promise<Result<Set<ModuleID>, Error>> {
        const getAllModulesArgMap: Map<string, boolean> = new Map<string,boolean>()
        const emptyGetModulesArg: GetModulesArgs = newGetModulesArgs(getAllModulesArgMap)

        const getModulesResponseResult = await this.backend.getModules(emptyGetModulesArg)
        if(getModulesResponseResult.isErr()){
            return err(getModulesResponseResult.error)
        }

        const modulesResponse = getModulesResponseResult.value

        const moduleIds: Set<ModuleID> = new Set<ModuleID>()

        modulesResponse.getModuleInfoMap().forEach((value: ModuleInfo, key: string) => {
            moduleIds.add(key)
        })

        return ok(moduleIds)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async uploadFiles(pathToArchive: string): Promise<Result<FilesArtifactID, Error>>  {
        const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(pathToArchive)
        if (archiverResponse.isErr()){
            return err(archiverResponse.error)
        }

        const args = newUploadFilesArtifactArgs(archiverResponse.value)
        const uploadResult = await this.backend.uploadFiles(args)
        if (uploadResult.isErr()){
            return err(uploadResult.error)
        }

        return ok(uploadResult.value.getUuid())
    }
      
    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async storeWebFiles(url: string): Promise<Result<FilesArtifactID, Error>> {
        const args = newStoreWebFilesArtifactArgs(url);
        const storeWebFilesArtifactResponseResult = await this.backend.storeWebFilesArtifact(args)
        if (storeWebFilesArtifactResponseResult.isErr()) {
            return err(storeWebFilesArtifactResponseResult.error)
        }
        const storeWebFilesArtifactResponse = storeWebFilesArtifactResponseResult.value;
        return ok(storeWebFilesArtifactResponse.getUuid())
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async storeServiceFiles(serviceId: ServiceID, absoluteFilepathOnServiceContainer: string): Promise<Result<FilesArtifactID, Error>> {
        const args = newStoreFilesArtifactFromServiceArgs(serviceId, absoluteFilepathOnServiceContainer)
        const storeFilesArtifactFromServiceResponseResult = await this.backend.storeFilesArtifactFromService(args)
        if (storeFilesArtifactFromServiceResponseResult.isErr()) {
            return err(storeFilesArtifactFromServiceResponseResult.error)
        }
        const storeFilesArtifactFromServiceResponse = storeFilesArtifactFromServiceResponseResult.value;
        return ok(storeFilesArtifactFromServiceResponse.getUuid())
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async pauseService(serviceId: string): Promise<Result<null, Error>> {
        const pauseServiceArgs: PauseServiceArgs = newPauseServiceArgs(serviceId)

        const pauseServiceResult = await this.backend.pauseService(pauseServiceArgs)
        if(pauseServiceResult.isErr()){
            return err(pauseServiceResult.error)
        }
        const pauseServiceResponse = pauseServiceResult.value
        return ok(null)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async unpauseService(serviceId: string): Promise<Result<null, Error>> {
        const unpauseServiceArgs: UnpauseServiceArgs = newUnpauseServiceArgs(serviceId)

        const unpauseServiceResult = await this.backend.unpauseService(unpauseServiceArgs)
        if(unpauseServiceResult.isErr()){
            return err(unpauseServiceResult.error)
        }
        const pauseServiceResponse = unpauseServiceResult.value
        return ok(null)
    }
  
    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
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
