/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

import * as jspb from "google-protobuf";
import {
    ExecCommandArgs,
    GetServicesArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    Port,
    StoreWebFilesArtifactArgs,
    ServiceInfo,
    GetServicesResponse,
    DownloadFilesArtifactArgs,
} from '../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { ServiceName } from './services/service';

// ==============================================================================================
//                           Shared Objects (Used By Multiple Endpoints)
// ==============================================================================================
export function newPort(number: number, transportProtocol: Port.TransportProtocol, maybeApplicationProtocol?: string) {
    const result: Port = new Port();
    result.setNumber(number);
    result.setTransportProtocol(transportProtocol);
    if (maybeApplicationProtocol) {
        result.setMaybeApplicationProtocol(maybeApplicationProtocol)
    }
    return result;
}

// ==============================================================================================
//                                       Get Services
// ==============================================================================================
export function newGetServicesArgs(serviceIdentifiers: Map<string, boolean>): GetServicesArgs{
    const result: GetServicesArgs = new GetServicesArgs();
    const resultServiceIdentifiersMap: jspb.Map<string, boolean> = result.getServiceIdentifiersMap()
    for (const [serviceName, booleanVal] of serviceIdentifiers) {
        resultServiceIdentifiersMap.set(serviceName, booleanVal);
    }

    return result;
}

export function newGetServicesResponse(serviceInfoMap: Map<string,ServiceInfo>): GetServicesResponse{
    const result: GetServicesResponse = new GetServicesResponse();
    const resultServiceMap: jspb.Map<string,ServiceInfo> = result.getServiceInfoMap()
    for (const [serviceName, serviceInfo] of serviceInfoMap) {
        resultServiceMap.set(serviceName, serviceInfo)
    }

    return result
}

// ==============================================================================================
//                                          Exec Command
// ==============================================================================================
export function newExecCommandArgs(setServiceIdentifier: ServiceName, command: string[]): ExecCommandArgs {
    const result: ExecCommandArgs = new ExecCommandArgs();
    result.setServiceIdentifier(setServiceIdentifier);
    result.setCommandArgsList(command);

    return result;
}

// ==============================================================================================
//                           Wait For Http Get Endpoint Availability
// ==============================================================================================
export function newWaitForHttpGetEndpointAvailabilityArgs(
        serviceIdentifier: string,
        port: number, 
        path: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpGetEndpointAvailabilityArgs {
    const result: WaitForHttpGetEndpointAvailabilityArgs = new WaitForHttpGetEndpointAvailabilityArgs();
    result.setServiceIdentifier(String(serviceIdentifier));
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
        serviceIdentifier: string,
        port: number, 
        path: string,
        requestBody: string,
        initialDelayMilliseconds: number, 
        retries: number, 
        retriesDelayMilliseconds: number, 
        bodyText: string): WaitForHttpPostEndpointAvailabilityArgs {
    const result: WaitForHttpPostEndpointAvailabilityArgs = new WaitForHttpPostEndpointAvailabilityArgs();
    result.setServiceIdentifier(serviceIdentifier);
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
//                                     Store Web Files Files
// ==============================================================================================
export function newStoreWebFilesArtifactArgs(url: string, name: string): StoreWebFilesArtifactArgs {
    const result: StoreWebFilesArtifactArgs = new StoreWebFilesArtifactArgs();
    result.setUrl(url);
    result.setName(name);
    return result;
}

// ==============================================================================================
//                                     Download Files
// ==============================================================================================
export function newDownloadFilesArtifactArgs(identifier: string): DownloadFilesArtifactArgs {
    const result: DownloadFilesArtifactArgs = new DownloadFilesArtifactArgs();
    result.setIdentifier(identifier);
    return result;
}
