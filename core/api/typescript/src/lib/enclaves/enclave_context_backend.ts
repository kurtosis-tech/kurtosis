import { Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import { 
    ExecuteBulkCommandsArgs, 
    GetModuleInfoArgs, 
    GetModulesResponse, 
    GetServiceInfoArgs, 
    GetServiceInfoResponse, 
    GetServicesResponse, 
    LoadModuleArgs, 
    RegisterFilesArtifactsArgs, 
    RegisterServiceArgs, 
    RegisterServiceResponse, 
    RemoveServiceArgs, 
    RepartitionArgs, 
    StartServiceArgs, 
    StartServiceResponse, 
    UnloadModuleArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientNode } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { EnclaveID } from "./enclave_context";

export default interface EnclaveContextBackend {
    getClient(): ApiContainerServiceClientWeb | ApiContainerServiceClientNode
    getEnclaveId(): EnclaveID
    loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<null, Error>>
    unloadModule(unloadModuleArgs: UnloadModuleArgs): Promise<Result<null,Error>>
    getModuleInfo(getModuleInfoArgs: GetModuleInfoArgs): Promise<Result<null, Error>>
    registerFilesArtifacts(registerFilesArtifactsArgs: RegisterFilesArtifactsArgs): Promise<Result<null,Error>>
    registerService(registerServiceArgs: RegisterServiceArgs): Promise<Result<RegisterServiceResponse, Error>>
    startService(startServiceArgs: StartServiceArgs): Promise<Result<StartServiceResponse, Error>>
    getServiceInfo(getServiceInfoArgs: GetServiceInfoArgs): Promise<Result<GetServiceInfoResponse, Error>>
    removeService(args: RemoveServiceArgs): Promise<Result<null, Error>>
    repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>>
    waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>>
    waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>>
    executeBulkCommands(executeBulkCommandsArgs: ExecuteBulkCommandsArgs): Promise<Result<null, Error>>
    getServices(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetServicesResponse, Error>>
    getModules(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetModulesResponse, Error>>
}