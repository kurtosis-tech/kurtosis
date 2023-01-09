/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import {ok, err, Result} from "neverthrow";
import log from "loglevel";
import { isNode as  isExecutionEnvNode} from "browser-or-node";
import * as jspb from "google-protobuf";
import type {
    Port,
    RemoveServiceArgs,
    ServiceConfig,
    GetServicesArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { GrpcNodeApiContainerClient } from "./grpc_node_api_container_client";
import { GrpcWebApiContainerClient } from "./grpc_web_api_container_client";
import type { GenericApiContainerClient } from "./generic_api_container_client";
import {
    newGetServicesArgs,
    newPort,
    newRemoveServiceArgs,
    newServiceConfig,
    newStartServicesArgs,
    newStoreWebFilesArtifactArgs,
    newUploadFilesArtifactArgs,
} from "../constructor_calls";
import type { ContainerConfig, FilesArtifactUUID } from "../services/container_config";
import type { ServiceID, ServiceGUID } from "../services/service";
import { ServiceContext } from "../services/service_context";
import { TransportProtocol, PortSpec } from "../services/port_spec";
import type { GenericPathJoiner } from "./generic_path_joiner";
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {
    ServiceInfo,
    StartServicesArgs,
    RunStarlarkScriptArgs,
    RunStarlarkPackageArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import * as path from "path";
import {parseKurtosisYaml} from "./kurtosis_yaml";
import {Readable} from "stream";
import {readStreamContentUntilClosed, StarlarkRunResult} from "./starlark_run_blocking";

export type EnclaveID = string;
export type PartitionID = string;

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)
const DEFAULT_PARTITION_ID: PartitionID = "";

export const KURTOSIS_YAML_FILENAME = "kurtosis.yml";


// Docs available at https://docs.kurtosis.com/sdk/#enclavecontext
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

    // Docs available at https://docs.kurtosis.com/sdk/#getenclaveid---enclaveid
    public getEnclaveId(): EnclaveID {
        return this.backend.getEnclaveId();
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptstring-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkScript(
        serializedStartosisScript: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<Readable, Error>> {
        const args = new RunStarlarkScriptArgs();
        args.setSerializedScript(serializedStartosisScript)
        args.setSerializedParams(serializedParams)
        args.setDryRun(dryRun)
        const scriptRunResult : Result<Readable, Error> = await this.backend.runStarlarkScript(args)
        if (scriptRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark script \n${scriptRunResult.error}`))
        }
        return ok(scriptRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptblockingstring-serializedstarlarkscript-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkScriptBlocking(
        serializedStartosisScript: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkScript(serializedStartosisScript, serializedParams, dryRun)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackagestring-packagerootpath-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkPackage(
        packageRootPath: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<Readable, Error>> {
        const args = await this.assembleRunStarlarkPackageArg(packageRootPath, serializedParams, dryRun)
        if (args.isErr()) {
            return err(new Error(`Unexpected error while assembling arguments to pass to the Starlark executor \n${args.error}`))
        }
        const packageRunResult : Result<Readable, Error> = await this.backend.runStarlarkPackage(args.value)
        if (packageRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark package \n${packageRunResult.error}`))
        }
        return ok(packageRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackageblockingstring-packagerootpath-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkPackageBlocking(
        packageRootPath: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkPackage(packageRootPath, serializedParams, dryRun)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runremotestarlarkpackagestring-packageid-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkRemotePackage(
        moduleId: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<Readable, Error>> {
        const args = new RunStarlarkPackageArgs();
        args.setPackageId(moduleId)
        args.setDryRun(dryRun)
        args.setSerializedParams(serializedParams)
        args.setRemote(true)
        const remotePackageRunResult : Result<Readable, Error> = await this.backend.runStarlarkPackage(args)
        if (remotePackageRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark package \n${remotePackageRunResult.error}`))
        }
        return ok(remotePackageRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackageblockingstring-packageid-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkRemotePackageBlocking(
        moduleId: string,
        serializedParams: string,
        dryRun: boolean,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkRemotePackage(moduleId, serializedParams, dryRun)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#addserviceserviceid-serviceid--containerconfig-containerconfig---servicecontext-servicecontext
    public async addService(
            serviceId: ServiceID,
            containerConfig: ContainerConfig
        ): Promise<Result<ServiceContext, Error>> {
        const containerConfigs : Map<ServiceID, ContainerConfig> = new Map<ServiceID, ContainerConfig>();
        containerConfigs.set(serviceId, containerConfig)
        const resultAddServiceToPartition : Result<[Map<ServiceID, ServiceContext>, Map<ServiceID, Error>], Error> = await this.addServicesToPartition(
            containerConfigs,
            DEFAULT_PARTITION_ID,
        );
        if (resultAddServiceToPartition.isErr()) {
            return err(resultAddServiceToPartition.error);
        }
        const [successfulServices, failedService] = resultAddServiceToPartition.value
        const serviceErr : Error | undefined = failedService.get(serviceId);
        if (serviceErr != undefined) {
            return err(new Error(`An error occurred adding service '${serviceId}' to the enclave in the default partition:\n${serviceErr}`))
        }
        const serviceCtx : ServiceContext | undefined = successfulServices.get(serviceId);
        if (serviceCtx == undefined){
            return err(new Error(`An error occurred retrieving the service context of service with ID ${serviceId} from result of adding service to partition. This should not happen and is a bug in Kurtosis.`))
        }
        return ok(serviceCtx);
    }

    // Docs available at https://docs.kurtosis.com/sdk/#addservicetopartitionserviceid-serviceid-partitionid-partitionid-containerconfig-containerconfig---servicecontext-servicecontext
    public async addServiceToPartition(
            serviceId: ServiceID,
            partitionId: PartitionID,
            containerConfig: ContainerConfig
        ): Promise<Result<ServiceContext, Error>> {
        const containerConfigs : Map<ServiceID, ContainerConfig> = new Map<ServiceID, ContainerConfig>();
        containerConfigs.set(serviceId, containerConfig)
        const resultAddServiceToPartition : Result<[Map<ServiceID, ServiceContext>, Map<ServiceID, Error>], Error> = await this.addServicesToPartition(
            containerConfigs,
            partitionId,
        );
        if (resultAddServiceToPartition.isErr()) {
            return err(resultAddServiceToPartition.error);
        }
        const [successfulServices, failedService] = resultAddServiceToPartition.value
        const serviceErr : Error | undefined = failedService.get(serviceId);
        if (serviceErr != undefined) {
            return err(new Error(`An error occurred adding service '${serviceId}' to the enclave in the default partition:\n${serviceErr}`))
        }
        const serviceCtx : ServiceContext | undefined = successfulServices.get(serviceId);
        if (serviceCtx == undefined){
            return err(new Error(`An error occurred retrieving the service context of service with ID ${serviceId} from result of adding service to partition. This should not happen and is a bug in Kurtosis.`))
        }
        return ok(serviceCtx);
    }

    // Docs available at https://docs.kurtosis.com/sdk/#addservicestopartitionmapserviceid-containerconfig-containerconfigs-partitionid-partitionid---mapserviceid-servicecontext-successfulservices-mapserviceid-error-failedservices
    public async addServicesToPartition(
        containerConfigs: Map<ServiceID, ContainerConfig>,
        partitionID: PartitionID,
    ): Promise<Result<[Map<ServiceID, ServiceContext>, Map<ServiceID, Error>], Error>> {
        const failedServicesPool: Map<ServiceID, Error> = new Map<ServiceID, Error>();
        const successfulServices: Map<ServiceID, ServiceContext> = new Map<ServiceID, ServiceContext>();

        const serviceConfigs = new Map<ServiceID, ServiceConfig>();
        for (const [serviceID, containerConfig] of containerConfigs.entries()) {
            log.trace(`Creating files artifact ID str -> mount dirpaths map for service with Id '${serviceID}'...`);
            const artifactIdStrToMountDirpath: Map<string, string> = new Map<string, string>();
            for (const [mountDirpath, filesArtifactId] of containerConfig.filesArtifactMountpoints) {
                artifactIdStrToMountDirpath.set(mountDirpath, filesArtifactId);
            }
            log.trace(`Successfully created files artifact ID str -> mount dirpaths map for service with Id '${serviceID}'`);

            const privatePorts = containerConfig.usedPorts;
            const privatePortsForApi: Map<string, Port> = new Map();
            for (const [portId, portSpec] of privatePorts.entries()) {
                const portSpecForApi: Port = newPort(
                    portSpec.number,
                    portSpec.transportProtocol,
                    portSpec.maybeApplicationProtocol
                )
                privatePortsForApi.set(portId, portSpecForApi);
            }
            //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
            const publicPorts = containerConfig.publicPorts;
            const publicPortsForApi: Map<string, Port> = new Map();
            for (const [portId, portSpec] of publicPorts.entries()) {
                const portSpecForApi: Port = newPort(
                    portSpec.number,
                    portSpec.transportProtocol,
                    portSpec.maybeApplicationProtocol
                )
                publicPortsForApi.set(portId, portSpecForApi);
            }
            //TODO finish the hack

            const serviceConfig: ServiceConfig = newServiceConfig(
                containerConfig.image,
                privatePortsForApi,
                publicPortsForApi,
                containerConfig.entrypointOverrideArgs,
                containerConfig.cmdOverrideArgs,
                containerConfig.environmentVariableOverrides,
                artifactIdStrToMountDirpath,
                containerConfig.cpuAllocationMillicpus,
                containerConfig.memoryAllocationMegabytes,
                containerConfig.privateIPAddrPlaceholder,
                partitionID,
            )
            serviceConfigs.set(serviceID, serviceConfig);
        }
        log.trace("Starting new services with Kurtosis API...");
        const startServicesArgs: StartServicesArgs = newStartServicesArgs(serviceConfigs)
        const startServicesResponseResult = await this.backend.startServices(startServicesArgs)
        if (startServicesResponseResult.isErr()) {
            return err(startServicesResponseResult.error)
        }
        const startServicesResponse = startServicesResponseResult.value;
        const successfulServicesInfo: jspb.Map<String, ServiceInfo> | undefined = startServicesResponse.getSuccessfulServiceIdsToServiceInfoMap();
        if (successfulServicesInfo === undefined) {
            return err(new Error("Expected StartServicesResponse to contain a field that does not exist."))
        }
        // defer-undo removes all successfully started services in case of errors in the future phases
        const shouldRemoveServices: Map<ServiceID, boolean> = new Map<ServiceID, boolean>();
        for (const [serviceIdStr, _] of successfulServicesInfo.entries()) {
            shouldRemoveServices.set(<ServiceID>serviceIdStr, true);
        }

        try {
            // Add services that failed to start to failed services pool
            const failedServices: jspb.Map<string, string> | undefined = startServicesResponse.getFailedServiceIdsToErrorMap();
            if (failedServices === undefined) {
                return err(new Error("Expected StartServicesResponse to contain a field that does not exist."))
            }
            for (const [serviceIdStr, serviceErrStr] of failedServices.entries()) {
                const serviceId: ServiceID = <ServiceID>serviceIdStr;
                failedServicesPool.set(serviceId, new Error(serviceErrStr))
            }
            for (const [serviceIdStr, serviceInfo] of successfulServicesInfo.entries()) {
                const serviceId: ServiceID = <ServiceID>serviceIdStr;
                const serviceCtxPrivatePorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
                    serviceInfo.getPrivatePortsMap(),
                );
                const serviceCtxPublicPorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
                    serviceInfo.getMaybePublicPortsMap(),
                );

                const serviceContext: ServiceContext = new ServiceContext(
                    this.backend,
                    serviceId,
                    serviceInfo.getServiceGuid(),
                    serviceInfo.getPrivateIpAddr(),
                    serviceCtxPrivatePorts,
                    serviceInfo.getMaybePublicIpAddr(),
                    serviceCtxPublicPorts,
                );
                successfulServices.set(serviceId, serviceContext)
                log.trace(`Successfully started service with ID '${serviceId}' with Kurtosis API`);
            }
            // Do not remove resources for successful services
            for (const [serviceId, _] of successfulServices) {
                shouldRemoveServices.delete(serviceId)
            }
        } finally {
            for (const[serviceId, _] of shouldRemoveServices) {
                // Do a best effort attempt to remove resources for this object to clean up after it failed
                // TODO: Migrate this to a bulk remove services call
                const removeServiceArgs : RemoveServiceArgs = newRemoveServiceArgs(serviceId)
                const removeServiceResult = await this.backend.removeService(removeServiceArgs);
                if (removeServiceResult.isErr()){
                    const errMsg = `"Attempted to remove service '${serviceId}' to delete its resources after it failed to start, but the following error occurred " +
                    "while attempting to remove the service:\n ${removeServiceResult.error}`
                    failedServicesPool.set(serviceId, new Error(errMsg))
                }
            }
        }
        return ok([successfulServices, failedServicesPool])
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getservicecontextserviceid-serviceid---servicecontext-servicecontext
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

        const serviceCtxPrivatePorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getPrivatePortsMap(),
        );
        const serviceCtxPublicPorts: Map<string, PortSpec> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getMaybePublicPortsMap(),
        );

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend,
            serviceId,
            serviceInfo.getServiceGuid(),
            serviceInfo.getPrivateIpAddr(),
            serviceCtxPrivatePorts,
            serviceInfo.getMaybePublicIpAddr(),
            serviceCtxPublicPorts,
        );

        return ok(serviceContext);
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getservices---mapserviceid--serviceguid-serviceids
    public async getServices(): Promise<Result<Map<ServiceID, ServiceGUID>, Error>> {
        const getAllServicesArgMap: Map<string, boolean> = new Map<string,boolean>()
        const emptyGetServicesArg: GetServicesArgs = newGetServicesArgs(getAllServicesArgMap)

        const getServicesResponseResult = await this.backend.getServices(emptyGetServicesArg)
        if(getServicesResponseResult.isErr()){
            return err(getServicesResponseResult.error)
        }

        const getServicesResponse = getServicesResponseResult.value

        const serviceInfos: Map<ServiceID, ServiceGUID> = new Map<ServiceID, ServiceGUID>()
        getServicesResponse.getServiceInfoMap().forEach((value: ServiceInfo, key: string) => {
            serviceInfos.set(key, value.getServiceGuid())
        });
        return ok(serviceInfos)
    }

    // Docs available at https://docs.kurtosis.com/sdk#uploadfilesstring-pathtoupload-string-artifactname
    public async uploadFiles(pathToArchive: string, name: string): Promise<Result<FilesArtifactUUID, Error>>  {
        const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(pathToArchive)
        if (archiverResponse.isErr()){
            return err(archiverResponse.error)
        }

        const args = newUploadFilesArtifactArgs(archiverResponse.value, name)
        const uploadResult = await this.backend.uploadFiles(args)
        if (uploadResult.isErr()){
            return err(uploadResult.error)
        }

        return ok(uploadResult.value.getUuid())
    }

    // Docs available at https://docs.kurtosis.com/sdk#storewebfilesstring-urltodownload-string-artifactname
    public async storeWebFiles(url: string, name: string): Promise<Result<FilesArtifactUUID, Error>> {
        const args = newStoreWebFilesArtifactArgs(url, name);
        const storeWebFilesArtifactResponseResult = await this.backend.storeWebFilesArtifact(args)
        if (storeWebFilesArtifactResponseResult.isErr()) {
            return err(storeWebFilesArtifactResponseResult.error)
        }
        const storeWebFilesArtifactResponse = storeWebFilesArtifactResponseResult.value;
        return ok(storeWebFilesArtifactResponse.getUuid())
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    private static convertApiPortsToServiceContextPorts(apiPorts: jspb.Map<string, Port>): Map<string, PortSpec> {
        const result: Map<string, PortSpec> = new Map();
        for (const [portId, apiPortSpec] of apiPorts.entries()) {
            const portProtocol: TransportProtocol = apiPortSpec.getTransportProtocol();
            const portNum: number = apiPortSpec.getNumber();
            const portSpec = new PortSpec(portNum, portProtocol, apiPortSpec.getMaybeApplicationProtocol());
            result.set(portId, portSpec)
        }
        return result;
    }

    private async assembleRunStarlarkPackageArg(packageRootPath: string, serializedParams: string, dryRun: boolean,): Promise<Result<RunStarlarkPackageArgs, Error>> {
        const kurtosisYamlFilepath = path.join(packageRootPath, KURTOSIS_YAML_FILENAME)

        const resultParseKurtosisYaml = await parseKurtosisYaml(kurtosisYamlFilepath)
        if (resultParseKurtosisYaml.isErr()) {
            return err(resultParseKurtosisYaml.error)
        }
        const kurtosisYaml = resultParseKurtosisYaml.value

        const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(packageRootPath)
        if (archiverResponse.isErr()){
            return err(archiverResponse.error)
        }

        const args = new RunStarlarkPackageArgs;
        args.setLocal(archiverResponse.value)
        args.setPackageId(kurtosisYaml.name)
        args.setSerializedParams(serializedParams)
        args.setDryRun(dryRun)
        return ok(args)
    }
}
