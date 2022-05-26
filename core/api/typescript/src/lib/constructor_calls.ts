/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import * as jspb from "google-protobuf";
import {
    ExecCommandArgs,
    GetServicesArgs,
    PartitionServices,
    PartitionConnections,
    PartitionConnectionInfo,
    RegisterServiceArgs,
    StartServiceArgs,
    RemoveServiceArgs,
    RepartitionArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    LoadModuleArgs,
    UnloadModuleArgs,
    ExecuteModuleArgs,
    GetModulesArgs,
    Port,
    StoreWebFilesArtifactArgs,
    StoreFilesArtifactFromServiceArgs,
    UploadFilesArtifactArgs,
    PauseServiceArgs,
    UnpauseServiceArgs,
    ModuleInfo,
    ServiceInfo,
    RemoveServiceResponse,
    UnloadModuleResponse,
    GetModulesResponse,
    GetServicesResponse
} from '../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { ServiceID } from './services/service';
import { PartitionID } from './enclaves/enclave_context';
import { ModuleID } from "./modules/module_context";

// ==============================================================================================
//                           Shared Objects (Used By Multiple Endpoints)
// ==============================================================================================
export function newPort(number: number, protocol: Port.Protocol) {
    const result: Port = new Port();
    result.setNumber(number);
    result.setProtocol(protocol);
    return result;
}


// ==============================================================================================
//                                     Load Module
// ==============================================================================================
export function newLoadModuleArgs(moduleId: ModuleID, image: string, serializedParams: string): LoadModuleArgs {
    const result: LoadModuleArgs = new LoadModuleArgs();
    result.setModuleId(String(moduleId));
    result.setContainerImage(image);
    result.setSerializedParams(serializedParams);

    return result;
}

// ==============================================================================================
//                                     Unload Module
// ==============================================================================================
export function newUnloadModuleArgs(moduleId: ModuleID): UnloadModuleArgs {
    const result: UnloadModuleArgs = new UnloadModuleArgs();
    result.setModuleId(String(moduleId));

    return result;
}

export function newUnloadModuleResponse(moduleGuid: string): UnloadModuleResponse {
    const result: UnloadModuleResponse = new UnloadModuleResponse();
    result.setModuleGuid(moduleGuid)

    return result;
}


// ==============================================================================================
//                                     Execute Module
// ==============================================================================================
export function newExecuteModuleArgs(moduleId: ModuleID, serializedParams: string): ExecuteModuleArgs {
    const result: ExecuteModuleArgs = new ExecuteModuleArgs();
    result.setModuleId(String(moduleId));
    result.setSerializedParams(serializedParams);

    return result;
}


// ==============================================================================================
//                                     Get Modules
// ==============================================================================================
export function newGetModulesArgs(moduleIds: Map<string, boolean>): GetModulesArgs {
    const result: GetModulesArgs = new GetModulesArgs();
    const moduleMap: jspb.Map<string, boolean> = result.getIdsMap();
    for (const [moduleId, isModuleIncluded] of moduleIds) {
        moduleMap.set(moduleId, isModuleIncluded);
    }

    return result;
}

export function newGetModulesResponse(moduleInfoMap: Map<string, ModuleInfo>): GetModulesResponse {
    const result: GetModulesResponse = new GetModulesResponse();
    const resultModuleInfoMap: jspb.Map<string, ModuleInfo> = result.getModuleInfoMap();
    for (const [moduleId, moduleInfo] of moduleInfoMap) {
        resultModuleInfoMap.set(moduleId, moduleInfo);
    }

    return result;
}


export function newModuleInfo(
    guid: string,
    privateIpAddr: string,
    privateGrpcPort: Port,
    maybePublicIpAddr: string,
    maybePublicGrpcPort: Port,): ModuleInfo {
    const result: ModuleInfo = new ModuleInfo();
    result.setGuid(guid)
    result.setPrivateIpAddr(privateIpAddr)
    result.setPrivateGrpcPort(privateGrpcPort)
    result.setMaybePublicIpAddr(maybePublicIpAddr)
    result.setMaybePublicGrpcPort(maybePublicGrpcPort)

    return result
}


// ==============================================================================================
//                                     Register Service
// ==============================================================================================
export function newRegisterServiceArgs(serviceId: ServiceID, partitionId: PartitionID): RegisterServiceArgs {
    const result: RegisterServiceArgs = new RegisterServiceArgs();
    result.setServiceId(String(serviceId));
    result.setPartitionId(String(partitionId));

    return result;
}


// ==============================================================================================
//                                        Start Service
// ==============================================================================================
export function newStartServiceArgs(
    serviceId: ServiceID,
    dockerImage: string,
    privatePorts: Map<string, Port>,
    publicPorts:  Map<string, Port>, //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    entrypointArgs: string[],
    cmdArgs: string[],
    dockerEnvVars: Map<string, string>,
    filesArtifactMountDirpaths: Map<string, string>,
): StartServiceArgs {
    const result: StartServiceArgs = new StartServiceArgs();
    result.setServiceId(String(serviceId));
    result.setDockerImage(dockerImage);
    const usedPortsMap: jspb.Map<string, Port> = result.getPrivatePortsMap();
    for (const [portId, portSpec] of privatePorts) {
        usedPortsMap.set(portId, portSpec);
    }
    const entrypointArgsArray: string[] = result.getEntrypointArgsList();
    for (const entryPoint of entrypointArgs) {
        entrypointArgsArray.push(entryPoint);
    }
    const cmdArgsArray: string[] = result.getCmdArgsList();
    for (const cmdArg of cmdArgs) {
        cmdArgsArray.push(cmdArg);
    }
    const dockerEnvVarArray: jspb.Map<string, string> = result.getDockerEnvVarsMap();
    for (const [name, value] of dockerEnvVars.entries()) {
        dockerEnvVarArray.set(name, value);
    }
    const filesArtificatMountDirpathsMap: jspb.Map<string, string> = result.getFilesArtifactMountpointsMap();
    for (const [artifactId, mountDirpath] of filesArtifactMountDirpaths.entries()) {
        filesArtificatMountDirpathsMap.set(artifactId, mountDirpath);
    }

    //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
    const publicPortsMap: jspb.Map<string, Port> = result.getPublicPortsMap();
    for (const [portId, portSpec] of publicPorts) {
        publicPortsMap.set(portId, portSpec);
    }
    //TODO finish the hack

    return result;
}

// ==============================================================================================
//                                       Get Services
// ==============================================================================================
export function newGetServicesArgs(serviceIds: Map<string, boolean>): GetServicesArgs{
    const result: GetServicesArgs = new GetServicesArgs();
    const resultServiceIdMap: jspb.Map<string, boolean> = result.getServiceIdsMap()
    for (const [serviceId, booleanVal] of serviceIds) {
        resultServiceIdMap.set(serviceId, booleanVal);
    }

    return result;
}

export function newGetServicesResponse(serviceInfoMap: Map<string,ServiceInfo>): GetServicesResponse{
    const result: GetServicesResponse = new GetServicesResponse();
    const resultServiceMap: jspb.Map<string,ServiceInfo> = result.getServiceInfoMap()
    for (const [serviceId, serviceInfo] of serviceInfoMap) {
        resultServiceMap.set(serviceId, serviceInfo)
    }

    return result
}

export function newServiceInfo(
    serviceGuid: string,
    privateIpAddr: string,
    privatePorts: Map<string, Port>,
    maybePublicIpAddr: string,
    maybePublicPorts: Map<string, Port>,
): ServiceInfo {
    const result: ServiceInfo = new ServiceInfo();
    result.setServiceGuid(serviceGuid)
    result.setMaybePublicIpAddr(maybePublicIpAddr)
    result.setPrivateIpAddr(privateIpAddr)

    const privatePortsMap: jspb.Map<string, Port> = result.getPrivatePortsMap()
    for (const [portName, privatePort] of privatePorts.entries()) {
        privatePortsMap.set(portName, privatePort)
    }
    const maybePublicPortsMap: jspb.Map<string, Port> = result.getMaybePublicPortsMap()
    for (const [portName, publicPort] of maybePublicPorts.entries()) {
        maybePublicPortsMap.set(portName, publicPort)
    }

    return result
}


// ==============================================================================================
//                                        Remove Service
// ==============================================================================================
export function newRemoveServiceArgs(serviceId: ServiceID, containerStopTimeoutSeconds: number): RemoveServiceArgs {
    const result: RemoveServiceArgs = new RemoveServiceArgs();
    result.setServiceId(serviceId);
    result.setContainerStopTimeoutSeconds(containerStopTimeoutSeconds);

    return result;
}

export function newRemoveServiceResponse(serviceGuid: string): RemoveServiceResponse {
    const result: RemoveServiceResponse = new RemoveServiceResponse();
    result.setServiceGuid(serviceGuid)
    return result
}


// ==============================================================================================
//                                          Repartition
// ==============================================================================================
export function newRepartitionArgs(
        partitionServices: Map<string, PartitionServices>, 
        partitionConns: Map<string, PartitionConnections>,
        defaultConnection: PartitionConnectionInfo): RepartitionArgs {
    const result: RepartitionArgs = new RepartitionArgs();
    const partitionServicesMap: jspb.Map<string, PartitionServices> = result.getPartitionServicesMap();
    for (const [partitionServiceId, partitionId] of partitionServices.entries()) {
        partitionServicesMap.set(partitionServiceId, partitionId);
    };
    const partitionConnsMap: jspb.Map<string, PartitionConnections> = result.getPartitionConnectionsMap();
    for (const [partitionConnId, partitionConn] of partitionConns.entries()) {
        partitionConnsMap.set(partitionConnId, partitionConn);
    };
    result.setDefaultConnection(defaultConnection);

    return result;
}

export function newPartitionServices(serviceIdStrSet: Set<string>): PartitionServices{
    const result: PartitionServices = new PartitionServices();
    const partitionServicesMap: jspb.Map<string, boolean> = result.getServiceIdSetMap();
    for (const serviceIdStr of serviceIdStrSet) {
        partitionServicesMap.set(serviceIdStr, true);
    }

    return result;
}


export function newPartitionConnections(allConnectionInfo: Map<string, PartitionConnectionInfo>): PartitionConnections {
    const result: PartitionConnections = new PartitionConnections();
    const partitionsMap: jspb.Map<string, PartitionConnectionInfo> = result.getConnectionInfoMap();
    for (const [partitionId, connectionInfo] of allConnectionInfo.entries()) {
        partitionsMap.set(partitionId, connectionInfo);
    }

    return result;
}

export function newPartitionConnectionInfo(packetLossPercentage: number): PartitionConnectionInfo {
    const partitionConnectionInfo: PartitionConnectionInfo = new PartitionConnectionInfo();
    partitionConnectionInfo.setPacketLossPercentage(packetLossPercentage);
    return partitionConnectionInfo;
}


// ==============================================================================================
//                                          Exec Command
// ==============================================================================================
export function newExecCommandArgs(serviceId: ServiceID, command: string[]): ExecCommandArgs {
    const result: ExecCommandArgs = new ExecCommandArgs();
    result.setServiceId(serviceId);
    result.setCommandArgsList(command);

    return result;
}

// ==============================================================================================
//                                          Pause/Unpause Service
// ==============================================================================================
export function newPauseServiceArgs(serviceId: ServiceID): PauseServiceArgs {
    const result: PauseServiceArgs = new PauseServiceArgs();
    result.setServiceId(serviceId);

    return result;
}

export function newUnpauseServiceArgs(serviceId: ServiceID): UnpauseServiceArgs {
    const result: UnpauseServiceArgs = new UnpauseServiceArgs();
    result.setServiceId(serviceId);

    return result;
}


// ==============================================================================================
//                           Wait For Http Get Endpoint Availability
// ==============================================================================================
export function newWaitForHttpGetEndpointAvailabilityArgs(
        serviceId: ServiceID,
        port: number, 
        path: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpGetEndpointAvailabilityArgs {
    const result: WaitForHttpGetEndpointAvailabilityArgs = new WaitForHttpGetEndpointAvailabilityArgs();
    result.setServiceId(String(serviceId));
    result.setPort(port);
    result.setPath(path);
    result.setInitialDelayMilliseconds(initialDelayMilliseconds);
    result.setRetries(retries);
    result.setRetriesDelayMilliseconds(retriesDelayMilliseconds);
    result.setBodyText(bodyText);

    return result;
}


// ==============================================================================================
//                           Wait For Http Post Endpoint Availability
// ==============================================================================================
export function newWaitForHttpPostEndpointAvailabilityArgs(
        serviceId: ServiceID,
        port: number, 
        path: string,
        requestBody: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpPostEndpointAvailabilityArgs {
    const result: WaitForHttpPostEndpointAvailabilityArgs = new WaitForHttpPostEndpointAvailabilityArgs();
    result.setServiceId(String(serviceId));
    result.setPort(port);
    result.setPath(path);
    result.setRequestBody(requestBody)
    result.setInitialDelayMilliseconds(initialDelayMilliseconds);
    result.setRetries(retries);
    result.setRetriesDelayMilliseconds(retriesDelayMilliseconds);
    result.setBodyText(bodyText);

    return result;
}

// ==============================================================================================
//                                     Download Files
// ==============================================================================================
export function newStoreWebFilesArtifactArgs(url: string): StoreWebFilesArtifactArgs {
    const result: StoreWebFilesArtifactArgs = new StoreWebFilesArtifactArgs();
    result.setUrl(url);
    return result;
}

// ==============================================================================================
//                             Store Files Artifact From Service
// ==============================================================================================
export function newStoreFilesArtifactFromServiceArgs(serviceId: string, sourcePath: string): StoreFilesArtifactFromServiceArgs {
    const result: StoreFilesArtifactFromServiceArgs = new StoreFilesArtifactFromServiceArgs();
    result.setServiceId(serviceId)
    result.setSourcePath(sourcePath)
    return result;
}

// ==============================================================================================
//                                      Upload Files
// ==============================================================================================
export function newUploadFilesArtifactArgs(data: Uint8Array) : UploadFilesArtifactArgs {
    const result: UploadFilesArtifactArgs = new UploadFilesArtifactArgs()
    result.setData(data)
    return result
}