import {ok, err, Result, Err} from "neverthrow";
import type { ServiceError } from "@grpc/grpc-js";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {
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
    StartServiceResponse,
    GetServicesResponse,
    LoadModuleArgs,
    UnloadModuleArgs,
    GetModuleInfoArgs,
    GetModuleInfoResponse,
    GetModulesResponse,
    ExecuteModuleResponse,
    ExecuteModuleArgs,
    ExecCommandArgs,
    ExecCommandResponse,
    UploadFilesArtifactArgs,
    UploadFilesArtifactResponse
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import type { ApiContainerServiceClient as ApiContainerServiceClientNode } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { GenericApiContainerClient } from "./generic_api_container_client";
import { EnclaveID } from "./enclave_context";
import "fs";
import "os";
import "path";

const COMPRESSION_EXTENSION = ".tgz"

export class GrpcNodeApiContainerClient implements GenericApiContainerClient {

    private readonly client: ApiContainerServiceClientNode;
    private readonly enclaveId: EnclaveID;

    constructor(client: ApiContainerServiceClientNode, enclaveId: EnclaveID) {
        this.client = client;
        this.enclaveId = enclaveId;
    }

    public getEnclaveId(): EnclaveID {
        return this.enclaveId;
    }

    public async loadModule(loadModuleArgs: LoadModuleArgs): Promise<Result<null, Error>> {
        const loadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.loadModule(loadModuleArgs, (error: ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

    public async unloadModule(unloadModuleArgs: UnloadModuleArgs): Promise<Result<null,Error>> {
        const unloadModulePromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.unloadModule(unloadModuleArgs, (error: ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

    public async getModuleInfo(getModuleInfoArgs: GetModuleInfoArgs): Promise<Result<null, Error>> {
        const getModuleInfoPromise: Promise<Result<GetModuleInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModuleInfo(getModuleInfoArgs, (error: ServiceError | null, response?: GetModuleInfoResponse) => {
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
        const getModuleInfoResponseResult: Result<GetModuleInfoResponse, Error> = await getModuleInfoPromise;
        if (getModuleInfoResponseResult.isErr()) {
            return err(getModuleInfoResponseResult.error);
        }

        return ok(null);
    }

    public async registerFilesArtifacts(registerFilesArtifactsArgs: RegisterFilesArtifactsArgs): Promise<Result<null,Error>> {
        const promiseRegisterFilesArtifacts: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerFilesArtifacts(registerFilesArtifactsArgs, (error: ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
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

    public async registerService(registerServiceArgs: RegisterServiceArgs): Promise<Result<RegisterServiceResponse, Error>>{
        const registerServicePromise: Promise<Result<RegisterServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.registerService(registerServiceArgs, (error: ServiceError | null, response?: RegisterServiceResponse) => {
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

    public async startService(startServiceArgs: StartServiceArgs): Promise<Result<StartServiceResponse, Error>>{
        const promiseStartService: Promise<Result<StartServiceResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.startService(startServiceArgs, (error: ServiceError | null, response?: StartServiceResponse) => {
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

    public async getServiceInfo(getServiceInfoArgs: GetServiceInfoArgs): Promise<Result<GetServiceInfoResponse, Error>> {
        const promiseGetServiceInfo: Promise<Result<GetServiceInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServiceInfo(getServiceInfoArgs, (error: ServiceError | null, response?: GetServiceInfoResponse) => {
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
            this.client.removeService(args, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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
            this.client.repartition(repartitionArgs, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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
            this.client.waitForHttpGetEndpointAvailability(availabilityArgs, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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
            this.client.waitForHttpPostEndpointAvailability(availabilityArgs, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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

    public async getServices(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetServicesResponse, Error>> {
        const promiseGetServices: Promise<Result<GetServicesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getServices(emptyArg, (error: ServiceError | null, response?: GetServicesResponse) => {
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

    public async getModules(emptyArg: google_protobuf_empty_pb.Empty): Promise<Result<GetModulesResponse, Error>> {
        const getModulesPromise: Promise<Result<GetModulesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getModules(emptyArg, (error: ServiceError | null, response?: GetModulesResponse) => {
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
            this.client.executeModule(executeModuleArgs, (error: ServiceError | null, response?: ExecuteModuleResponse) => {
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
            this.client.execCommand(execCommandArgs, (error: ServiceError | null, response?: ExecCommandResponse) => {
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

        const execCommandResponse = execCommandResponseResult.value;
        return ok(execCommandResponse)
    }

    public async uploadFiles(uploadFilesArtifactArgs: UploadFilesArtifactArgs): Promise<Result<UploadFilesArtifactResponse, Error>> {
        const uploadFilesPromise: Promise<Result<UploadFilesArtifactResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.uploadFilesArtifact(uploadFilesArtifactArgs, (error: ServiceError | null, response?: UploadFilesArtifactResponse) => {
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
        const uploadFilesResult: Result<UploadFilesArtifactResponse, Error> = await uploadFilesPromise;
        if(uploadFilesResult.isErr()){
            return err(uploadFilesResult.error)
        }

        const uploadFilesResponse = uploadFilesResult.value
        return ok(uploadFilesResponse)
    }

    private TarCreator(sourcePath : string) {
        const targz = require("targz")
        const filesystem = require("fs")
        const os = require("os")
        const path = require("path")

        //Check if it exists
        if(!filesystem.existsSync(sourcePath)) {
            return err(new Error("The file or folder you want to upload does not exist."))
        }

        //Make directory for usage.
        var absoluteTarPath : string = ""
        var tempDirectoryErr : Error | null = null
        filesystem.mkdtemp(os.tmpdir(),  (tempDirError : Error, folder: string) => {
            tempDirectoryErr = tempDirError
            absoluteTarPath = path.join(folder, path.basename(sourcePath)) ;
        });

        if (tempDirectoryErr != null){
            return err(tempDirectoryErr)
        }

        const baseName = path.basename(sourcePath) + COMPRESSION_EXTENSION
        const options  = {
            src: sourcePath,
            dest: path.join(absoluteTarPath,baseName),
        }

        var error : Error | string | null = null
        targz.compress(options, function(compressErr: Error) { error = compressErr })
        if(error != null){
            return err(error)
        }

        if (filesystem.existsSync(options.dest)){
            return err(new Error(`Your files were compressed but could not be found at ${options.dest}.`))
        }

        //this.client.uploadFilesArtifact()
        //check to see if context had error
        //return uuid
    }
}