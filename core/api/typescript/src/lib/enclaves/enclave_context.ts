/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import { ApiContainerServiceClient } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import {
    RegisterFilesArtifactsArgs,
    PortBinding,
    RegisterServiceArgs,
    RegisterServiceResponse,
    StartServiceArgs,
    GetServiceInfoArgs,
    GetServiceInfoResponse,
    RemoveServiceArgs,
    PartitionConnectionInfo,
    PartitionServices,
    PartitionConnections,
    RepartitionArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    ExecuteBulkCommandsArgs,
    StartServiceResponse,
    GetServicesResponse,
    LoadModuleArgs,
    UnloadModuleArgs,
    GetModuleInfoArgs,
    GetModuleInfoResponse,
    GetModulesResponse,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ModuleID, ModuleContext } from "../modules/module_context";
import { ServiceID} from "../services/service";
import { SharedPath } from "../services/shared_path";
import { ServiceContext} from "../services/service_context";
import {
    newLoadModuleArgs,
    newGetModuleInfoArgs,
    newRegisterFilesArtifactsArgs,
    newRegisterServiceArgs,
    newStartServiceArgs,
    newGetServiceInfoArgs,
    newRemoveServiceArgs,
    newPartitionServices,
    newPartitionConnections,
    newRepartitionArgs,
    newWaitForHttpGetEndpointAvailabilityArgs,
    newWaitForHttpPostEndpointAvailabilityArgs,
    newExecuteBulkCommandsArgs,
    newUnloadModuleArgs
} from "../constructor_calls";
import { ok, err, Result } from "neverthrow";
import * as log from "loglevel";
import * as grpc from "grpc";
import * as path from "path"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import { ContainerConfig, FilesArtifactID } from "../services/container_config";

export type EnclaveID = string;

export type PartitionID = string;

// This will always resolve to the default partition ID (regardless of whether such a partition exists in the enclave,
//  or it was repartitioned away)
const DEFAULT_PARTITION_ID: PartitionID = "";

// The path on the user service container where the enclave data dir will be bind-mounted
const SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT: string = "/kurtosis-enclave-data";

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
export class EnclaveContext {
    private readonly client: ApiContainerServiceClient;

    private readonly enclaveId: EnclaveID;
    
    // The location on the filesystem where this code is running where the enclave data dir exists
    private readonly enclaveDataDirpath: string;

    /*
    Creates a new EnclaveContext object with the given parameters.
    */
    constructor(client: ApiContainerServiceClient, enclaveId: EnclaveID, enclaveDataDirpath: string) {
        this.client = client;
        this.enclaveId = enclaveId;
        this.enclaveDataDirpath = enclaveDataDirpath;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public getEnclaveId(): EnclaveID {
        return this.enclaveId;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async loadModule(
            moduleId: ModuleID,
            image: string,
            serializedParams: string): Promise<Result<ModuleContext, Error>> {
        const args: LoadModuleArgs = newLoadModuleArgs(moduleId, image, serializedParams);

        const loadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.loadModule(args, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        const loadModuleResult: Result<google_protobuf_empty_pb.Empty, Error> = await loadModulePromise;
        if (!loadModuleResult.isOk()) {
            return err(loadModuleResult.error);
        }

        const moduleCtx: ModuleContext = new ModuleContext(this.client, moduleId);
        return ok(moduleCtx);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async unloadModule(moduleId: ModuleID): Promise<Result<null,Error>> {
        const args: UnloadModuleArgs = newUnloadModuleArgs(moduleId);

        const unloadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.unloadModule(args, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        })

        const unloadModuleResult: Result<google_protobuf_empty_pb.Empty, Error> = await unloadModulePromise;
        if (!unloadModuleResult.isOk()) {
            return err(unloadModuleResult.error);
        }
        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>> {
        const args: GetModuleInfoArgs = newGetModuleInfoArgs(moduleId);
        
        const getModuleInfoPromise: Promise<Result<GetModuleInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModuleInfo(args, (error: grpc.ServiceError | null, response?: GetModuleInfoResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                    }
                    resolve(ok(response!));
                } else {
                    resolve(err(error));
                }
            })
        });
        const getModuleInfoResult: Result<GetModuleInfoResponse, Error> = await getModuleInfoPromise;
        if (!getModuleInfoResult.isOk()) {
            return err(getModuleInfoResult.error);
        }

        const moduleCtx: ModuleContext = new ModuleContext(this.client, moduleId);
        return ok(moduleCtx);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async registerFilesArtifacts(filesArtifactUrls: Map<FilesArtifactID, string>): Promise<Result<null,Error>> {
        const filesArtifactIdStrsToUrls: Map<string, string> = new Map();
        for (const [artifactId, url] of filesArtifactUrls.entries()) {
            filesArtifactIdStrsToUrls.set(String(artifactId), url);
        }
        const args: RegisterFilesArtifactsArgs = newRegisterFilesArtifactsArgs(filesArtifactIdStrsToUrls);
        
        const promiseRegisterFilesArtifacts: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerFilesArtifacts(args, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        const resultRegisterFilesArtifacts: Result<google_protobuf_empty_pb.Empty, Error> = await promiseRegisterFilesArtifacts;
        if (!resultRegisterFilesArtifacts.isOk()) {
            return err(resultRegisterFilesArtifacts.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async addService(
            serviceId: ServiceID,
            containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error>
        ): Promise<Result<[ServiceContext, Map<string, PortBinding>], Error>> {

        const resultAddServiceToPartition: Result<[ServiceContext, Map<string, PortBinding>], Error> = await this.addServiceToPartition(
            serviceId,
            DEFAULT_PARTITION_ID,
            containerConfigSupplier,
        );

        if (!resultAddServiceToPartition.isOk()) {
            return err(resultAddServiceToPartition.error);
        }

        return ok(resultAddServiceToPartition.value);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async addServiceToPartition(
            serviceId: ServiceID,
            partitionId: PartitionID,
            containerConfigSupplier: (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error>
        ): Promise<Result<[ServiceContext, Map<string, PortBinding>], Error>> {

        log.trace("Registering new service ID with Kurtosis API...");
        const registerServiceArgs: RegisterServiceArgs = newRegisterServiceArgs(serviceId, partitionId);

        const promiseRegisterService: Promise<Result<RegisterServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerService(registerServiceArgs, (error: grpc.ServiceError | null, response?: RegisterServiceResponse) => {
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
        const resultRegisterService: Result<RegisterServiceResponse, Error> = await promiseRegisterService;
        if (!resultRegisterService.isOk()) {
            return err(resultRegisterService.error);
        }
        const registerServiceResp: RegisterServiceResponse = resultRegisterService.value;

        log.trace("New service successfully registered with Kurtosis API");

        const serviceIpAddr: string = registerServiceResp.getIpAddr();
        const relativeServiceDirpath: string = registerServiceResp.getRelativeServiceDirpath();

        const sharedDirectory = this.getSharedDirectory(relativeServiceDirpath)

        log.trace("Generating container config object using the container config supplier...")
        const containerConfigSupplierResult: Result<ContainerConfig, Error> = containerConfigSupplier(serviceIpAddr, sharedDirectory);
        if (!containerConfigSupplierResult.isOk()){
            return err(containerConfigSupplierResult.error);
        }
        const containerConfig: ContainerConfig = containerConfigSupplierResult.value;
        log.trace("Container config object successfully generated")

        log.trace("Creating files artifact ID str -> mount dirpaths map...");
        const artifactIdStrToMountDirpath: Map<string, string> = new Map();
        for (const [filesArtifactId, mountDirpath] of containerConfig.getFilesArtifactMountpoints().entries()) {

            artifactIdStrToMountDirpath.set(String(filesArtifactId), mountDirpath);
        }
        log.trace("Successfully created files artifact ID str -> mount dirpaths map");

        log.trace("Starting new service with Kurtosis API...");
        const startServiceArgs: StartServiceArgs = newStartServiceArgs(
            serviceId, 
            containerConfig.getImage(), 
            containerConfig.getUsedPortsSet(),
            containerConfig.getEntrypointOverrideArgs(),
            containerConfig.getCmdOverrideArgs(),
            containerConfig.getEnvironmentVariableOverrides(),
            SERVICE_ENCLAVE_DATA_DIR_MOUNTPOINT,
            artifactIdStrToMountDirpath);

        const promiseStartService: Promise<Result<StartServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.startService(startServiceArgs, (error: Error | null, response?: StartServiceResponse) => {
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
        const resultStartService: Result<StartServiceResponse, Error> = await promiseStartService;
        if (!resultStartService.isOk()) {
            return err(resultStartService.error);
        }

        log.trace("Successfully started service with Kurtosis API");

        const serviceContext: ServiceContext = new ServiceContext(
            this.client,
            serviceId,
            serviceIpAddr,
            sharedDirectory);

        const resp: StartServiceResponse = resultStartService.value;
        const resultMap: Map<string, PortBinding> = new Map();
        for (const [key, value] of resp.getUsedPortsHostPortBindingsMap().entries()) {
            resultMap.set(key, value);
        }
        return ok<[ServiceContext, Map<string, PortBinding>], Error>([serviceContext, resultMap]);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async getServiceContext(serviceId: ServiceID): Promise<Result<ServiceContext, Error>> {
        const getServiceInfoArgs: GetServiceInfoArgs = newGetServiceInfoArgs(serviceId);
        
        const promiseGetServiceInfo: Promise<Result<GetServiceInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServiceInfo(getServiceInfoArgs, (error: Error | null, response?: GetServiceInfoResponse) => {
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
        const resultGetServiceInfo: Result<GetServiceInfoResponse, Error> = await promiseGetServiceInfo;
        if (!resultGetServiceInfo.isOk()) {
            return err(resultGetServiceInfo.error);
        }

        const serviceResponse: GetServiceInfoResponse = resultGetServiceInfo.value;
        if (serviceResponse.getIpAddr() === "") {
            return err(new Error(
                "Kurtosis API reported an empty IP address for service " + serviceId +  " - this should never happen, and is a bug with Kurtosis!",
                ) 
            );
        }

        const relativeServiceDirpath: string = serviceResponse.getRelativeServiceDirpath();
        if (relativeServiceDirpath === "") {
            return err(new Error(
                "Kurtosis API reported an empty relative service directory path for service " + serviceId + " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const enclaveDataDirMountDirpathOnSvcContainer: string = serviceResponse.getEnclaveDataDirMountDirpath();
        if (enclaveDataDirMountDirpathOnSvcContainer === "") {
            return err(new Error(
                "Kurtosis API reported an empty enclave data dir mount dirpath for service " + serviceId + " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const sharedDirectory: SharedPath = this.getSharedDirectory(relativeServiceDirpath)

        const serviceContext: ServiceContext = new ServiceContext(
            this.client,
            serviceId,
            serviceResponse.getIpAddr(),
            sharedDirectory,
        );

        return ok(serviceContext);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async removeService(serviceId: ServiceID, containerStopTimeoutSeconds: number): Promise<Result<null, Error>> {

        log.debug("Removing service '" + serviceId + "'...");
        // NOTE: This is kinda weird - when we remove a service we can never get it back so having a container
        //  stop timeout doesn't make much sense. It will make more sense when we can stop/start containers
        // Independent of adding/removing them from the enclave
        const args: RemoveServiceArgs = newRemoveServiceArgs(serviceId, containerStopTimeoutSeconds);
        
        const removeServicePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.removeService(args, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultRemoveService: Result<null, Error> = await removeServicePromise;
        if (!resultRemoveService.isOk()) {
            return err(resultRemoveService.error);
        }

        log.debug("Successfully removed service ID " + serviceId);

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async repartitionNetwork(
            partitionServices: Map<PartitionID, Set<ServiceID>>,
            partitionConnections: Map<PartitionID, Map<PartitionID, PartitionConnectionInfo>>,
            defaultConnection: PartitionConnectionInfo): Promise<Result<null, Error>> {

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
            for (const [partitionBId, connInfo] of partitionAConnsMap.entries()) {

                const partitionBIdStr: string = String(partitionBId);
                partitionAConnsStrMap.set(partitionBIdStr, connInfo);
            }
            const partitionAConns: PartitionConnections = newPartitionConnections(partitionAConnsStrMap);
            const partitionAIdStr: string = String(partitionAId);
            reqPartitionConns.set(partitionAIdStr, partitionAConns);
        }

        const repartitionArgs: RepartitionArgs = newRepartitionArgs(reqPartitionServices, reqPartitionConns, defaultConnection);

        const promiseRepartition: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.repartition(repartitionArgs, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultRepartition: Result<null, Error> = await promiseRepartition;
        if (!resultRepartition.isOk()) {
            return err(resultRepartition.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async waitForHttpGetEndpointAvailability(
        serviceId: ServiceID,
        port: number, 
        path: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): Promise<Result<null, Error>> {
    const availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs = newWaitForHttpGetEndpointAvailabilityArgs(
        serviceId,
        port,
        path,
        initialDelayMilliseconds,
        retries,
        retriesDelayMilliseconds,
        bodyText);

    const promiseWaitForHttpGetEndpointAvailability: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
        this.client.waitForHttpGetEndpointAvailability(availabilityArgs, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
            if (error === null) {
                resolve(ok(null));
            } else {
                resolve(err(error));
            }
        })
    });
    const resultWaitForHttpGetEndpointAvailability: Result<null, Error> = await promiseWaitForHttpGetEndpointAvailability;
    if (!resultWaitForHttpGetEndpointAvailability.isOk()) {
        return err(resultWaitForHttpGetEndpointAvailability.error);
    }

    return ok(null);
}

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
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

        const promiseWaitForHttpPostEndpointAvailability: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.waitForHttpPostEndpointAvailability(availabilityArgs, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultWaitForHttpPostEndpointAvailability: Result<null, Error> = await promiseWaitForHttpPostEndpointAvailability;
        if (!resultWaitForHttpPostEndpointAvailability.isOk()) {
            return err(resultWaitForHttpPostEndpointAvailability.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async executeBulkCommands(bulkCommandsJson: string): Promise<Result<null, Error>> {

        const args: ExecuteBulkCommandsArgs = newExecuteBulkCommandsArgs(bulkCommandsJson);
        
        const promiseExecuteBulkCommands: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeBulkCommands(args, (error: Error | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultExecuteBulkCommands: Result<null, Error> = await promiseExecuteBulkCommands;
        if (!resultExecuteBulkCommands.isOk()) {
            return err(resultExecuteBulkCommands.error);
        }

        return ok(null);
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async getServices(): Promise<Result<Set<ServiceID>, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        
        const promiseGetServices: Promise<Result<GetServicesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServices(emptyArg, (error: Error | null, response?: GetServicesResponse) => {
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

        const resultGetServices: Result<GetServicesResponse, Error> = await promiseGetServices;
        if (!resultGetServices.isOk()) {
            return err(resultGetServices.error);
        }

        const getServicesResponse: GetServicesResponse = resultGetServices.value;

        const serviceIDs: Set<ServiceID> = new Set<ServiceID>()

        getServicesResponse.getServiceIdsMap().forEach((value: boolean, key: string) => {
            serviceIDs.add(key)
        });

        return ok(serviceIDs)
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async getModules(): Promise<Result<Set<ModuleID>, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        
        const getModulesPromise: Promise<Result<GetModulesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModules(emptyArg, (error: Error | null, response?: GetModulesResponse) => {
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

        const getModulesResult: Result<GetModulesResponse, Error> = await getModulesPromise;
        if (!getModulesResult.isOk()) {
            return err(getModulesResult.error);
        }

        const getModulesResponse: GetModulesResponse = getModulesResult.value;

        const moduleIds: Set<ModuleID> = new Set<ModuleID>()

        getModulesResponse.getModuleIdsMap().forEach((value: boolean, key: string) => {
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
}
