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
    ExecCommandArgs,
    ExecCommandResponse,
    PauseServiceArgs,
    UnpauseServiceArgs,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse,
    StoreWebFilesArtifactResponse,
    StoreWebFilesArtifactArgs,
    GetServicesArgs,
    RemoveServiceResponse,
    RunStarlarkScriptArgs,
    RunStarlarkPackageArgs,
    StarlarkRunResponseLine,
    DownloadFilesArtifactArgs,
    DownloadFilesArtifactResponse,
    GetExistingAndHistoricalServiceIdentifiersResponse, ListFilesArtifactNamesAndUuidsResponse,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { GenericApiContainerClient } from "./generic_api_container_client";
import { EnclaveUUID } from "./enclave_context";
import {Readable} from "stream";

export class GrpcWebApiContainerClient implements GenericApiContainerClient {

    private readonly client: ApiContainerServiceClientWeb;
    private readonly enclaveUuid: EnclaveUUID;
    private readonly enclaveName: string;

    constructor(client: ApiContainerServiceClientWeb, enclaveUuid: EnclaveUUID, enclaveName: string) {
        this.client = client;
        this.enclaveUuid = enclaveUuid;
        this.enclaveName = enclaveName;
    }

    public getEnclaveUuid():EnclaveUUID {
        return this.enclaveUuid;
    }

    public getEnclaveName(): string {
        return this.enclaveName;
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

    public async downloadFilesArtifact(downloadFilesArtifactArgs: DownloadFilesArtifactArgs): Promise<Result<DownloadFilesArtifactResponse, Error>> {
        const downloadFilesArtifactPromise: Promise<Result<DownloadFilesArtifactResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.downloadFilesArtifact(downloadFilesArtifactArgs, {}, (error: grpc_web.RpcError | null, response?: DownloadFilesArtifactResponse) => {
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

        const downloadFilesArtifactResponseResult: Result<DownloadFilesArtifactResponse, Error> = await downloadFilesArtifactPromise;
        if(downloadFilesArtifactResponseResult.isErr()){
            return err(downloadFilesArtifactResponseResult.error)
        }

        const downloadFilesArtifactResponse = downloadFilesArtifactResponseResult.value;
        return ok(downloadFilesArtifactResponse)
    }

    public async getExistingAndHistoricalServiceIdentifiers(): Promise<Result<GetExistingAndHistoricalServiceIdentifiersResponse, Error>>{
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        const getExistingAndHistoricalServiceIdentifiersPromise: Promise<Result<GetExistingAndHistoricalServiceIdentifiersResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getExistingAndHistoricalServiceIdentifiers(emptyArg, {},(error: grpc_web.RpcError | null, response?: GetExistingAndHistoricalServiceIdentifiersResponse) => {
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

        const getExistingAndHistoricalServiceIdentifiersResult: Result<GetExistingAndHistoricalServiceIdentifiersResponse, Error> = await getExistingAndHistoricalServiceIdentifiersPromise;
        if (getExistingAndHistoricalServiceIdentifiersResult.isErr()) {
            return err(getExistingAndHistoricalServiceIdentifiersResult.error)
        }

        return ok(getExistingAndHistoricalServiceIdentifiersResult.value);
    }

    public async getAllFilesArtifactNamesAndUuids(): Promise<Result<ListFilesArtifactNamesAndUuidsResponse, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        const getAllFilesArtifactNamesAndUuidsPromise: Promise<Result<ListFilesArtifactNamesAndUuidsResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.listFilesArtifactNamesAndUuids(emptyArg, {},(error: grpc_web.RpcError | null, response?: ListFilesArtifactNamesAndUuidsResponse) => {
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

        const getAllFilesArtifactNamesAndUuidsResult: Result<ListFilesArtifactNamesAndUuidsResponse, Error> = await getAllFilesArtifactNamesAndUuidsPromise;
        if (getAllFilesArtifactNamesAndUuidsResult.isErr()) {
            return err(getAllFilesArtifactNamesAndUuidsResult.error)
        }

        return ok(getAllFilesArtifactNamesAndUuidsResult.value);
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
