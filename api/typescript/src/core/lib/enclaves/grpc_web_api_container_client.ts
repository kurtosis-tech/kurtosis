import {ok, err, Result} from "neverthrow";
import * as grpc_web from "grpc-web";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {
    StartServicesArgs,
    StartServicesResponse,
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
    RenderTemplatesToFilesArtifactArgs,
    RenderTemplatesToFilesArtifactResponse,
    RunStarlarkScriptArgs,
    RunStarlarkPackageArgs,
    StarlarkRunResponseLine,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { GenericApiContainerClient } from "./generic_api_container_client";
import { EnclaveID } from "./enclave_context";
import {Readable} from "stream";

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

    public async loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<null, Error>> {
        const loadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.loadModule(loadModuleArgs, {}, (error: grpc_web.RpcError | null, response?: google_protobuf_empty_pb.Empty) => {
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
        const loadModulePromiseResult: Result<google_protobuf_empty_pb.Empty, Error> = await loadModulePromise;
        if (loadModulePromiseResult.isErr()) {
            return err(loadModulePromiseResult.error);
        }

        return ok(null);
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

    public async runStarlarkScript(serializedStarlarkScript: RunStarlarkScriptArgs): Promise<Result<Readable, Error>> {
        const promiseRunStarlarkScript: Promise<Result<grpc_web.ClientReadableStream<StarlarkRunResponseLine>, Error>> = new Promise((resolve, _unusedReject) => {
            resolve(ok(this.client.runStarlarkScript(serializedStarlarkScript, {})));
        })
        const runStarlarkScriptResult: Result<grpc_web.ClientReadableStream<StarlarkRunResponseLine>, Error> = await promiseRunStarlarkScript;
        if (runStarlarkScriptResult.isErr()) {
            return err(runStarlarkScriptResult.error)
        }
        const starlarkExecutionResponseLinesReadable: Readable = this.forwardStarlarkRunResponseLinesStreamToReadable(runStarlarkScriptResult.value)
        return ok(starlarkExecutionResponseLinesReadable)
    }

    public async runStarlarkPackage(starlarkPackageArgs: RunStarlarkPackageArgs): Promise<Result<Readable, Error>> {
        const promiseRunStarlarkPackage: Promise<Result<grpc_web.ClientReadableStream<StarlarkRunResponseLine>, Error>> = new Promise((resolve, _unusedReject) => {
            resolve(ok(this.client.runStarlarkPackage(starlarkPackageArgs)))
        })

        const runStarlarkPackageResult: Result<grpc_web.ClientReadableStream<StarlarkRunResponseLine>, Error> = await promiseRunStarlarkPackage;
        if (runStarlarkPackageResult.isErr()) {
            return err(runStarlarkPackageResult.error)
        }
        const starlarkRunResponseLinesReadable: Readable = this.forwardStarlarkRunResponseLinesStreamToReadable(runStarlarkPackageResult.value)
        return ok(starlarkRunResponseLinesReadable)
    }

    public async startServices(startServicesArgs: StartServicesArgs): Promise<Result<StartServicesResponse, Error>>{
        const promiseStartServices: Promise<Result<StartServicesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.startServices(startServicesArgs, {}, (error: grpc_web.RpcError | null, response?: StartServicesResponse) => {
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
        const resultStartServices: Result<StartServicesResponse, Error> = await promiseStartServices;
        if (resultStartServices.isErr()) {
            return err(resultStartServices.error);
        }

        const startServicesResponse: StartServicesResponse = resultStartServices.value;
        return ok(startServicesResponse)
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

    public async renderTemplatesToFilesArtifact(renderTemplatesToFilesArtifactArgs: RenderTemplatesToFilesArtifactArgs): Promise<Result<RenderTemplatesToFilesArtifactResponse, Error>> {
        const renderTemplatesToFilesArtifactPromise: Promise<Result<RenderTemplatesToFilesArtifactResponse, Error>> = new Promise( (resolve, _unusedReject) => {
            this.client.renderTemplatesToFilesArtifact(renderTemplatesToFilesArtifactArgs, {}, (error: grpc_web.RpcError | null, response?: RenderTemplatesToFilesArtifactResponse) => {
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
        const renderTemplatesToFilesArtifactResponseResult: Result<RenderTemplatesToFilesArtifactResponse, Error> = await renderTemplatesToFilesArtifactPromise;
        if (renderTemplatesToFilesArtifactResponseResult.isErr()) {
            return err(renderTemplatesToFilesArtifactResponseResult.error)
        }
        const renderTemplatesToFilesArtifactResponse = renderTemplatesToFilesArtifactResponseResult.value;
        return ok(renderTemplatesToFilesArtifactResponse);
    }

    private forwardStarlarkRunResponseLinesStreamToReadable(incomingStream: grpc_web.ClientReadableStream<StarlarkRunResponseLine>): Readable {
        const starlarkExecutionResponseLinesReadable: Readable = new Readable({
            objectMode: true, //setting object mode is to allow pass objects in the readable.push() method
            read() {
            } //this is mandatory to implement, we implement empty as it's describe in the implementation examples here: https://nodesource.com/blog/understanding-streams-in-nodejs/
        })
        starlarkExecutionResponseLinesReadable.on("close", function () {
            //Cancel the GRPC stream when users close the source stream
            incomingStream.cancel();
        })

        incomingStream.on('data', responseLine => {
            starlarkExecutionResponseLinesReadable.push(responseLine)
        })
        incomingStream.on('error', err => {
            if (!starlarkExecutionResponseLinesReadable.destroyed) {
                //Propagate the GRPC error to the service logs readable
                const grpcStreamErr = new Error(`An error has been returned from the kurtosis execution response lines stream. Error:\n ${err}`)
                starlarkExecutionResponseLinesReadable.emit('error', grpcStreamErr);
            }
        })
        incomingStream.on('end', () => {
            //Emit streams 'end' event when the GRPC stream has end
            if (!starlarkExecutionResponseLinesReadable.destroyed) {
                starlarkExecutionResponseLinesReadable.emit('end');
            }
        })
        return starlarkExecutionResponseLinesReadable
    }
}
