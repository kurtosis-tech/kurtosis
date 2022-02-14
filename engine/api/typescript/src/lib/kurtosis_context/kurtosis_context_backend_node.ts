import * as grpc_node from "@grpc/grpc-js";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {err, ok, Result} from "neverthrow";
import { ApiContainerServiceClientNode } from "kurtosis-core-api-lib";
import {
    EngineServiceClientNode,
    CleanArgs,
    CleanResponse,
    CreateEnclaveArgs,
    CreateEnclaveResponse,
    DestroyEnclaveArgs,
    EnclaveAPIContainerHostMachineInfo,
    GetEnclavesResponse,
    GetEngineInfoResponse,
    StopEnclaveArgs
} from "../../index";
import {newCleanArgs, newDestroyEnclaveArgs, newStopEnclaveArgs} from "../constructor_calls";
import { KurtosisContextBackend } from "./kurtosis_context";

type EnclaveID = string;

export class GrpcNodeKurtosisContextBackend implements KurtosisContextBackend {
    private readonly client: EngineServiceClientNode

    constructor(client: EngineServiceClientNode){
        this.client = client
    }

    public async getEngineInfo(): Promise<Result<GetEngineInfoResponse, Error>> {
        const getEngineInfoPromise: Promise<Result<GetEngineInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

            this.client.getEngineInfo(emptyArg, (error: grpc_node.ServiceError | null, response?: GetEngineInfoResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this should never " + "happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    if(error.code === grpc_node.status.UNAVAILABLE){
                        resolve(err(new Error("The Kurtosis Engine Server is unavailable and is probably not running; you " +
                            "will need to start it using the Kurtosis CLI before you can create a connection to it")));
                    }
                    resolve(err(error));
                }
            })
        });


        const getEngineInfoResult: Result<GetEngineInfoResponse, Error> = await getEngineInfoPromise;
        if (getEngineInfoResult.isErr()) {
            return err(getEngineInfoResult.error)
        }

        const engineInfoResponse: GetEngineInfoResponse = getEngineInfoResult.value;

        return ok(engineInfoResponse)
    }

    public async createEnclaveResponse(args: CreateEnclaveArgs): Promise<Result<CreateEnclaveResponse, Error>> {
        const createEnclavePromise: Promise<Result<CreateEnclaveResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.createEnclave(args, {}, (error: grpc_node.ServiceError | null, response?: CreateEnclaveResponse) => {
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

        const createEnclaveResult: Result<CreateEnclaveResponse, Error> = await createEnclavePromise;
        if (createEnclaveResult.isErr()) {
            return err(createEnclaveResult.error)
        }

        const enclaveResponse: CreateEnclaveResponse = createEnclaveResult.value;
        
        return ok(enclaveResponse)

    }

     public async stopEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>> {
        const args: StopEnclaveArgs = newStopEnclaveArgs(enclaveId)

        const stopEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.stopEnclave(args, (error: grpc_node.ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const stopEnclaveResult: Result<null, Error> = await stopEnclavePromise;
        if (stopEnclaveResult.isErr()) {
            return err(stopEnclaveResult.error);
        }

        return ok(null);
    }

    public async destroyEnclave(enclaveId: EnclaveID): Promise<Result<null, Error>> {
        const args: DestroyEnclaveArgs = newDestroyEnclaveArgs(enclaveId);

        const destroyEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.destroyEnclave(args, (error: grpc_node.ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
                if (error === null) {
                    resolve(ok(null));
                } else {
                    resolve(err(error));
                }
            })
        });
        const destroyEnclaveResult: Result<null, Error> = await destroyEnclavePromise;
        if (destroyEnclaveResult.isErr()) {
            return err(destroyEnclaveResult.error);
        }

        return ok(null);
    }

    public async clean( shouldCleanAll : boolean): Promise<Result<Set<string>, Error>>{

        const cleanArgs: CleanArgs = newCleanArgs(shouldCleanAll);

        const cleanPromise: Promise<Result<CleanResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.clean(cleanArgs, (error: grpc_node.ServiceError | null, response?: CleanResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error("No error was encountered but the response was still falsy; this " +
                            "should never happen")));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    resolve(err(error));
                }
            })
        });
        const cleanResult: Result<CleanResponse, Error> = await cleanPromise;
        if (cleanResult.isErr()) {
            return err(cleanResult.error)
        }
        const cleanResponse: CleanResponse = cleanResult.value;

        const result: Set<string> = new Set();
        for (let enclaveID of cleanResponse.getRemovedEnclaveIdsMap().keys()) {
            result.add(enclaveID);
        }
        return ok(result);
    }

    public createApiClient(localhostIpAddress:string, apiContainerHostMachineInfo:EnclaveAPIContainerHostMachineInfo):Result<ApiContainerServiceClientNode,Error>{
        const apiContainerHostMachineGrpcUrl: string = `${localhostIpAddress}:${apiContainerHostMachineInfo.getGrpcPortOnHostMachine()}`

        let apiContainerClient: ApiContainerServiceClientNode;
        
        try {
            apiContainerClient = new ApiContainerServiceClientNode(apiContainerHostMachineGrpcUrl, grpc_node.ChannelCredentials.createInsecure());
        } catch(exception) {
            if (exception instanceof Error) {
                return err(exception);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that" +
                " wasn't an error: " + exception
            ));
        }

        return ok(apiContainerClient)

    }

    public async getEnclavesResponse(): Promise<Result<GetEnclavesResponse, Error>>{
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getEnclavesPromise: Promise<Result<GetEnclavesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEnclaves(emptyArg, (error: grpc_node.ServiceError | null, response?: GetEnclavesResponse) => {
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
        const getEnclavesResponseResult: Result<GetEnclavesResponse, Error> = await getEnclavesPromise;
        if (!getEnclavesResponseResult.isOk()) {
            return err(getEnclavesResponseResult.error)
        }

        return ok(getEnclavesResponseResult.value);
    }
}