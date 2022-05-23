import {ok, err, Result, Err} from "neverthrow";
import * as grpc_web from "grpc-web";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {
    RegisterServiceArgs,
    RegisterServiceResponse,
    StartServiceArgs,
    RemoveServiceArgs,
    RepartitionArgs,
    WaitForHttpGetEndpointAvailabilityArgs,
    WaitForHttpPostEndpointAvailabilityArgs,
    GetServicesResponse,
    LoadModuleArgs,
    UnloadModuleArgs,
    GetModulesResponse,
    ExecuteModuleArgs,
    ExecuteModuleResponse,
    ExecCommandArgs,
    ExecCommandResponse,
    PauseServiceArgs,
    UnpauseServiceArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    StoreWebFilesArtifactResponse,
    StoreWebFilesArtifactArgs,
    StoreFilesArtifactFromServiceArgs,
    StoreFilesArtifactFromServiceResponse,
    GetServicesArgs,
    GetModulesArgs,
    UnloadModuleResponse,
    RemoveServiceResponse,
    ServiceInfo, ModuleInfo,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { GenericApiContainerClient } from "./generic_api_container_client";
import { EnclaveID } from "./enclave_context";

export class GrpcWebApiContainerClient implements GenericApiContainerClient {

    private readonly client: ApiContainerServiceClientWeb;
    private readonly enclaveId: EnclaveID;

    constructor(client: ApiContainerServiceClientWeb, enclaveId: EnclaveID) {
        this.client = client;
        this.enclaveId = enclaveId;
    }

    public getEnclaveId():EnclaveID {
        return this.enclaveId;
    }

    public async loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<ModuleInfo, Error>> {
        const loadModulePromise: Promise<Result<ModuleInfo, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.loadModule(loadModuleArgs, {}, (error: grpc_web.RpcError | null, response?: ModuleInfo) => {
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
        const loadModulePromiseResult: Result<ModuleInfo, Error> = await loadModulePromise;
        if (loadModulePromiseResult.isErr()) {
            return err(loadModulePromiseResult.error);
        }
        const loadModuleResponse = loadModulePromiseResult.value

        return ok(loadModuleResponse);
    }

    public async unloadModule(unloadModuleArgs: UnloadModuleArgs): Promise<Result<UnloadModuleResponse,Error>> {
        const unloadModulePromise: Promise<Result<UnloadModuleResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.unloadModule(unloadModuleArgs, {}, (error: grpc_web.RpcError | null, response?: UnloadModuleResponse | undefined) => {
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

        const unloadModuleResult: Result<UnloadModuleResponse, Error> = await unloadModulePromise;
        if (unloadModuleResult.isErr()) {
            return err(unloadModuleResult.error);
        }

        return ok(unloadModuleResult.value);
    }

    public async registerService(registerServiceArgs: RegisterServiceArgs): Promise<Result<RegisterServiceResponse, Error>>{
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

        const registerServiceResponse = registerServicePromiseResult.value;
        return ok(registerServiceResponse)
    }

    public async startService(startServiceArgs: StartServiceArgs): Promise<Result<ServiceInfo, Error>>{
        const promiseStartService: Promise<Result<ServiceInfo, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.startService(startServiceArgs, {}, (error: grpc_web.RpcError | null, response?: ServiceInfo) => {
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
        const resultStartService: Result<ServiceInfo, Error> = await promiseStartService;
        if (resultStartService.isErr()) {
            return err(resultStartService.error);
        }

        const startServiceResponse: ServiceInfo = resultStartService.value;
        return ok(startServiceResponse)
    }

    public async removeService(args: RemoveServiceArgs): Promise<Result<RemoveServiceResponse, Error>> {
        const removeServicePromise: Promise<Result<RemoveServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.removeService(args, {}, (error: grpc_web.RpcError | null, response?: RemoveServiceResponse | undefined) => {
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
        const resultRemoveService: Result<RemoveServiceResponse, Error> = await removeServicePromise;
        if (resultRemoveService.isErr()) {
            return err(resultRemoveService.error);
        }
        return ok(resultRemoveService.value);
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

    public async waitForHttpGetEndpointAvailability(availabilityArgs: WaitForHttpGetEndpointAvailabilityArgs): Promise<Result<null, Error>> {
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

    public async waitForHttpPostEndpointAvailability(availabilityArgs: WaitForHttpPostEndpointAvailabilityArgs): Promise<Result<null, Error>> {
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

    public async getServices(getServicesArgs: GetServicesArgs): Promise<Result<GetServicesResponse, Error>> {
        const promiseGetServices: Promise<Result<GetServicesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServices(getServicesArgs, {}, (error: grpc_web.RpcError | null, response?: GetServicesResponse) => {
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

        const getServicesResponse = resultGetServices.value;
        return ok(getServicesResponse)
    }

    public async getModules(getModulesArgs: GetModulesArgs): Promise<Result<GetModulesResponse, Error>> {
        const getModulesPromise: Promise<Result<GetModulesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModules(getModulesArgs, {}, (error: grpc_web.RpcError | null, response?: GetModulesResponse) => {
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

        const getModulesResponse = getModulesResult.value;
        return ok(getModulesResponse)
    }

    public async executeModule(executeModuleArgs: ExecuteModuleArgs): Promise<Result<ExecuteModuleResponse, Error>> {
        const executeModulePromise: Promise<Result<ExecuteModuleResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeModule(executeModuleArgs, {}, (error: grpc_web.RpcError | null, response?: ExecuteModuleResponse) => {
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
        const executeModuleResult: Result<ExecuteModuleResponse, Error> = await executeModulePromise;
        if (executeModuleResult.isErr()) {
            return err(executeModuleResult.error);
        }
        
        const executeModuleResponse: ExecuteModuleResponse = executeModuleResult.value;
        return ok(executeModuleResponse);
    }

    public async execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>> {
        const execCommandPromise: Promise<Result<ExecCommandResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.execCommand(execCommandArgs, {}, (error: grpc_web.RpcError | null, response?: ExecCommandResponse) => {
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
        const execCommandResponseResult: Result<ExecCommandResponse, Error> = await execCommandPromise;
        if(execCommandResponseResult.isErr()){
            return err(execCommandResponseResult.error)
        }

        const execCommandResponse = execCommandResponseResult.value
        return ok(execCommandResponse)
    }

    public async pauseService(pauseServiceArgs: PauseServiceArgs): Promise<Result<null, Error>> {
        const pauseServicePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.pauseService(pauseServiceArgs,  {}, (error: grpc_web.RpcError | null) => {
                if (error === null) {
                    resolve(ok(null))
                } else {
                    resolve(err(error));
                }
            })
        });
        const pauseServiceResult: Result<null, Error> = await pauseServicePromise;
        if(pauseServiceResult.isErr()){
            return err(pauseServiceResult.error)
        }
        return ok(null)
    }

    public async unpauseService(unpauseServiceArgs: UnpauseServiceArgs): Promise<Result<null, Error>> {
        const unpauseServicePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.unpauseService(unpauseServiceArgs, {}, (error: grpc_web.RpcError | null) => {
                if (error === null) {
                    resolve(ok(null))
                } else {
                    resolve(err(error));
                }
            })
        });
        const unpauseServiceResult: Result<null, Error> = await unpauseServicePromise;
        if (unpauseServiceResult.isErr()) {
            return err(unpauseServiceResult.error)
        }

        return ok(null)
    }

    public async uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>> {
        const uploadFilesArtifactPromise: Promise<Result<UploadFilesArtifactResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.uploadFilesArtifact(uploadFilesArtifactArgs, {}, (error: grpc_web.RpcError | null, response?: UploadFilesArtifactResponse) => {
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
        const uploadFilesArtifactResponseResult = await uploadFilesArtifactPromise;
        if(uploadFilesArtifactResponseResult.isErr()){
            return err(uploadFilesArtifactResponseResult.error)
        }

        const uploadFilesArtifactResponse = uploadFilesArtifactResponseResult.value
        return ok(uploadFilesArtifactResponse)
    }

    public async storeWebFilesArtifact(storeWebFilesArtifactArgs: StoreWebFilesArtifactArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>> {
        const storeWebFilesArtifactPromise: Promise<Result<StoreWebFilesArtifactResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.storeWebFilesArtifact(storeWebFilesArtifactArgs, {}, (error: grpc_web.RpcError | null, response?: StoreWebFilesArtifactResponse) => {
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
        const storeWebFilesArtifactResponseResult = await storeWebFilesArtifactPromise;
        if (storeWebFilesArtifactResponseResult.isErr()) {
            return err(storeWebFilesArtifactResponseResult.error)
        }
        const storeWebFilesArtifactResponse = storeWebFilesArtifactResponseResult.value;
        return ok(storeWebFilesArtifactResponse);
    }

    public async storeFilesArtifactFromService(storeFilesArtifactFromServiceArgs: StoreFilesArtifactFromServiceArgs): Promise<Result<StoreWebFilesArtifactResponse, Error>> {
        const storeFilesArtifactFromServicePromise: Promise<Result<StoreFilesArtifactFromServiceResponse, Error>> = new Promise( (resolve, _unusedReject) => {
            this.client.storeFilesArtifactFromService(storeFilesArtifactFromServiceArgs, {}, (error: grpc_web.RpcError | null, response?: StoreFilesArtifactFromServiceResponse) => {
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
        const storeFilesArtifactFromServiceResponseResult: Result<StoreFilesArtifactFromServiceResponse, Error>  = await storeFilesArtifactFromServicePromise;
        if (storeFilesArtifactFromServiceResponseResult.isErr()) {
            return err(storeFilesArtifactFromServiceResponseResult.error)
        }
        const storeFilesArtifactFromServiceResponse = storeFilesArtifactFromServiceResponseResult.value;
        return ok(storeFilesArtifactFromServiceResponse);
    }
}