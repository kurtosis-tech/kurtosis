import {
    ExecCommandArgs,
    RegisterFilesArtifactsArgs,
    GetServiceInfoArgs,
    PartitionServices,
    PartitionConnections,
    PartitionConnectionInfo,
    RegisterServiceArgs,
    StartServiceArgs,
    RemoveServiceArgs,
    RepartitionArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    ExecuteBulkCommandsArgs,
    LoadModuleArgs,
    UnloadModuleArgs,
    ExecuteModuleArgs,
    GetModuleInfoArgs
} from '../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { ServiceID } from './services/service';
import { PartitionID } from './enclaves/enclave_context';
import { ModuleID } from "./modules/module_context";
import * as jspb from "google-protobuf";

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
//                                     Get Module Info
// ==============================================================================================
export function newGetModuleInfoArgs(moduleId: ModuleID): GetModuleInfoArgs {
    const result: GetModuleInfoArgs = new GetModuleInfoArgs();
    result.setModuleId(String(moduleId));

    return result;
}


// ==============================================================================================
//                                       Register Files Artifacts
// ==============================================================================================
export function newRegisterFilesArtifactsArgs(filesArtifactIdStrsToUrls: Map<string, string>): RegisterFilesArtifactsArgs {
    const result: RegisterFilesArtifactsArgs = new RegisterFilesArtifactsArgs();
    const filesArtifactUrlsMap: jspb.Map<string, string> = result.getFilesArtifactUrlsMap();
    for (const [artifactId, artifactUrl] of filesArtifactIdStrsToUrls.entries()) {
        filesArtifactUrlsMap.set(artifactId, artifactUrl);
    }
    return result;
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
        usedPorts: Set<string>,
        entrypointArgs: string[],
        cmdArgs: string[],
        dockerEnvVars: Map<string, string>,
        enclaveDataDirMntDirpath: string,
        filesArtifactMountDirpaths: Map<string, string>): StartServiceArgs {
    const result: StartServiceArgs = new StartServiceArgs();
    result.setServiceId(String(serviceId));
    result.setDockerImage(dockerImage);
    const usedPortsMap: jspb.Map<string, boolean> = result.getUsedPortsMap();
    for (const portId of usedPorts) {
        usedPortsMap.set(portId, true);
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
    result.setEnclaveDataDirMntDirpath(enclaveDataDirMntDirpath);
    const filesArtificatMountDirpathsMap: jspb.Map<string, string> = result.getFilesArtifactMountDirpathsMap();
    for (const [artifactId, mountDirpath] of filesArtifactMountDirpaths.entries()) {
        filesArtificatMountDirpathsMap.set(artifactId, mountDirpath);
    }

    return result;
}

// ==============================================================================================
//                                       Get Service Info
// ==============================================================================================
export function newGetServiceInfoArgs(serviceId: ServiceID): GetServiceInfoArgs{
    const result: GetServiceInfoArgs = new GetServiceInfoArgs();
    result.setServiceId(String(serviceId));

    return result;
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
//                                      Execute Bulk Commands
// ==============================================================================================
export function newExecuteBulkCommandsArgs(serializedCommands: string): ExecuteBulkCommandsArgs {
    const result: ExecuteBulkCommandsArgs = new ExecuteBulkCommandsArgs();
    result.setSerializedCommands(serializedCommands);

    return result;
}

