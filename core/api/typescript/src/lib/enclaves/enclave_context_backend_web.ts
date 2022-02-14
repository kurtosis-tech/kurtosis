import { ok, err, Result } from "neverthrow";
import * as grpc_web from "grpc-web";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import { 
    ApiContainerServiceClientWeb, 
    ModuleID, 
    ModuleContext,
    ServiceID,
    FilesArtifactID,
    RegisterFilesArtifactsArgs,
    RegisterServiceArgs,
    RegisterServiceResponse,
    StartServiceArgs,
    GetServiceInfoArgs,
    GetServiceInfoResponse,
    RemoveServiceArgs,
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
    newLoadModuleArgs,
    newGetModuleInfoArgs,
    newRegisterFilesArtifactsArgs,
    newRegisterServiceArgs,
    newGetServiceInfoArgs,
    newWaitForHttpGetEndpointAvailabilityArgs,
    newWaitForHttpPostEndpointAvailabilityArgs,
    newExecuteBulkCommandsArgs,
    newUnloadModuleArgs,
} from "../../index";
import { EnclaveContextBackend } from "./enclave_context";

export type EnclaveID = string;

export type PartitionID = string;

export class GrpcWebEnclaveContextBackend implements EnclaveContextBackend {

    public readonly client: ApiContainerServiceClientWeb;

    private readonly enclaveId: EnclaveID;

    constructor(client: ApiContainerServiceClientWeb, enclaveId: EnclaveID) {
        this.client = client;
        this.enclaveId = enclaveId;
    }

    public getEnclaveId(): EnclaveID {
        return this.enclaveId;
    }

    public async loadModule(
            moduleId: ModuleID,
            image: string,
            serializedParams: string
        ): Promise<Result<ModuleContext, Error>> {
            const args: LoadModuleArgs = newLoadModuleArgs(moduleId, image, serializedParams);

            const loadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                this.client.loadModule(args, {}, (error: grpc_web.RpcError | null, response?: google_protobuf_empty_pb.Empty) => {
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
            if (loadModuleResult.isErr()) {
                return err(loadModuleResult.error);
            }

            const moduleCtx: ModuleContext = new ModuleContext(this.client, moduleId);
            return ok(moduleCtx);
    }

    public async unloadModule(moduleId: ModuleID): Promise<Result<null,Error>> {
        const args: UnloadModuleArgs = newUnloadModuleArgs(moduleId);

        const unloadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.unloadModule(args, {}, (error: grpc_web.RpcError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        if (unloadModuleResult.isErr()) {
            return err(unloadModuleResult.error);
        }
        return ok(null);
    }

    public async getModuleContext(moduleId: ModuleID): Promise<Result<ModuleContext, Error>> {
        const args: GetModuleInfoArgs = newGetModuleInfoArgs(moduleId);
        
        const getModuleInfoPromise: Promise<Result<GetModuleInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModuleInfo(args, {}, (error: grpc_web.RpcError | null, response?: GetModuleInfoResponse) => {
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
        if (getModuleInfoResult.isErr()) {
            return err(getModuleInfoResult.error);
        }

        const moduleCtx: ModuleContext = new ModuleContext(this.client, moduleId);
        return ok(moduleCtx);
    }

    public async registerFilesArtifacts(filesArtifactUrls: Map<FilesArtifactID, string>): Promise<Result<null,Error>> {
        const filesArtifactIdStrsToUrls: Map<string, string> = new Map();
        for (const [artifactId, url] of filesArtifactUrls.entries()) {
            filesArtifactIdStrsToUrls.set(String(artifactId), url);
        }
        const args: RegisterFilesArtifactsArgs = newRegisterFilesArtifactsArgs(filesArtifactIdStrsToUrls);
        
        const promiseRegisterFilesArtifacts: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerFilesArtifacts(args, {}, (error: grpc_web.RpcError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        if (resultRegisterFilesArtifacts.isErr()) {
            return err(resultRegisterFilesArtifacts.error);
        }

        return ok(null);
    }

    public async registerService( serviceId: ServiceID, partitionId: PartitionID): Promise<Result<RegisterServiceResponse, Error>>{
        const registerServiceArgs: RegisterServiceArgs = newRegisterServiceArgs(serviceId, partitionId);

        const registerServicePromise: Promise<Result<RegisterServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerService(registerServiceArgs, {}, (error: grpc_web.RpcError | null, response?: RegisterServiceResponse) => {
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
        const registerServicePromiseResult: Result<RegisterServiceResponse, Error> = await registerServicePromise;
        if (registerServicePromiseResult.isErr()) {
            return err(registerServicePromiseResult.error);
        }

        const registerServiceResponse: RegisterServiceResponse = registerServicePromiseResult.value;

        return ok(registerServiceResponse)
    }

    public async startService(startServiceArgs: StartServiceArgs): Promise<Result<StartServiceResponse, Error>>{
        const promiseStartService: Promise<Result<StartServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.startService(startServiceArgs, {}, (error: grpc_web.RpcError | null, response?: StartServiceResponse) => {
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
        if (resultStartService.isErr()) {
            return err(resultStartService.error);
        }
        const startServiceResponse: StartServiceResponse = resultStartService.value;

        return ok(startServiceResponse)
    }

    public async getServiceInfo(serviceId: ServiceID): Promise<Result<GetServiceInfoResponse, Error>> {
        const getServiceInfoArgs: GetServiceInfoArgs = newGetServiceInfoArgs(serviceId);
        
        const promiseGetServiceInfo: Promise<Result<GetServiceInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServiceInfo(getServiceInfoArgs, {}, (error: grpc_web.RpcError | null, response?: GetServiceInfoResponse) => {
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
        if (resultGetServiceInfo.isErr()) {
            return err(resultGetServiceInfo.error);
        }
        const getServiceInfoResponse: GetServiceInfoResponse = resultGetServiceInfo.value;

       return ok(getServiceInfoResponse)
    }

    public async removeService(args: RemoveServiceArgs): Promise<Result<null, Error>> {
        const removeServicePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.removeService(args, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultRemoveService: Result<null, Error> = await removeServicePromise;
        if (resultRemoveService.isErr()) {
            return err(resultRemoveService.error);
        }
        return ok(null);
    }

    public async repartitionNetwork(repartitionArgs: RepartitionArgs): Promise<Result<null, Error>> {
        const promiseRepartition: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.repartition(repartitionArgs, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultRepartition: Result<null, Error> = await promiseRepartition;
        if (resultRepartition.isErr()) {
            return err(resultRepartition.error);
        }

        return ok(null);
    }

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
            bodyText);

        const promiseWaitForHttpGetEndpointAvailability: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.waitForHttpGetEndpointAvailability(availabilityArgs, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultWaitForHttpGetEndpointAvailability: Result<null, Error> = await promiseWaitForHttpGetEndpointAvailability;
        if (resultWaitForHttpGetEndpointAvailability.isErr()) {
            return err(resultWaitForHttpGetEndpointAvailability.error);
        }

        return ok(null);
    }

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

        const promiseWaitForHttpPostEndpointAvailability: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.waitForHttpPostEndpointAvailability(availabilityArgs, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultWaitForHttpPostEndpointAvailability: Result<null, Error> = await promiseWaitForHttpPostEndpointAvailability;
        if (resultWaitForHttpPostEndpointAvailability.isErr()) {
            return err(resultWaitForHttpPostEndpointAvailability.error);
        }

        return ok(null);
    }

    public async executeBulkCommands(bulkCommandsJson: string): Promise<Result<null, Error>> {
        const args: ExecuteBulkCommandsArgs = newExecuteBulkCommandsArgs(bulkCommandsJson);
        
        const promiseExecuteBulkCommands: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeBulkCommands(args, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const resultExecuteBulkCommands: Result<null, Error> = await promiseExecuteBulkCommands;
        if (resultExecuteBulkCommands.isErr()) {
            return err(resultExecuteBulkCommands.error);
        }

        return ok(null);
    }

    public async getServices(): Promise<Result<GetServicesResponse, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        
        const promiseGetServices: Promise<Result<GetServicesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServices(emptyArg, {}, (error: grpc_web.RpcError | null, response?: GetServicesResponse) => {
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
        if (resultGetServices.isErr()) {
            return err(resultGetServices.error);
        }

        const getServicesResponse: GetServicesResponse = resultGetServices.value;

        return ok(getServicesResponse)
    }

    public async getModules(): Promise<Result<GetModulesResponse, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        
        const getModulesPromise: Promise<Result<GetModulesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModules(emptyArg, {}, (error: grpc_web.RpcError | null, response?: GetModulesResponse) => {
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
        if (getModulesResult.isErr()) {
            return err(getModulesResult.error);
        }

        const getModulesResponse: GetModulesResponse = getModulesResult.value;

        return ok(getModulesResponse)
    }
}