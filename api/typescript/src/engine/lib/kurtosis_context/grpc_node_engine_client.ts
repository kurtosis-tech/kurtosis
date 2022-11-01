import {err, ok, Result} from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import type {EngineServiceClient as EngineServiceClientNode} from "../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_pb";
import type {GenericEngineClient} from "./generic_engine_client";
import type {
    CleanArgs,
    CleanResponse,
    CreateEnclaveArgs,
    CreateEnclaveResponse,
    DestroyEnclaveArgs,
    GetEnclavesResponse,
    GetEngineInfoResponse,
    GetUserServiceLogsArgs,
    GetUserServiceLogsResponse,
    StopEnclaveArgs
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import {LogLine} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import {NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG} from "../consts";
import type {ClientReadableStream, ServiceError} from "@grpc/grpc-js";
import * as jspb from "google-protobuf";
import {ServiceGUID} from "../../../core/lib/services/service";
import {Readable} from "stream";

export class GrpcNodeEngineClient implements GenericEngineClient {
    private readonly client: EngineServiceClientNode

    constructor(client: EngineServiceClientNode) {
        this.client = client
    }

    public async getEngineInfo(): Promise<Result<GetEngineInfoResponse, Error>> {
        const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")

        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getEngineInfoPromise: Promise<Result<GetEngineInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEngineInfo(emptyArg, (error: ServiceError | null, response?: GetEngineInfoResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)));
                    } else {
                        resolve(ok(response!));
                    }
                } else {

                    if (error.code === grpc_node.status.UNAVAILABLE) {
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
            this.client.createEnclave(args, {}, (error: ServiceError | null, response?: CreateEnclaveResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)));
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

    public async stopEnclave(stopEnclaveArgs: StopEnclaveArgs): Promise<Result<null, Error>> {
        const stopEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.stopEnclave(stopEnclaveArgs, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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

    public async destroyEnclave(destroyEnclaveArgs: DestroyEnclaveArgs): Promise<Result<null, Error>> {
        const destroyEnclavePromise: Promise<Result<null, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.destroyEnclave(destroyEnclaveArgs, {}, (error: ServiceError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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

    public async clean(cleanArgs: CleanArgs): Promise<Result<CleanResponse, Error>> {
        const cleanPromise: Promise<Result<CleanResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.clean(cleanArgs, (error: ServiceError | null, response?: CleanResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)));
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
            return err(cleanResult.error);
        }

        const cleanResponse: CleanResponse = cleanResult.value;
        return ok(cleanResponse);
    }

    public async getEnclavesResponse(): Promise<Result<GetEnclavesResponse, Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        const getEnclavesPromise: Promise<Result<GetEnclavesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEnclaves(emptyArg, (error: ServiceError | null, response?: GetEnclavesResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    resolve(err(error));
                }
            })
        });

        const getEnclavesResponseResult: Result<GetEnclavesResponse, Error> = await getEnclavesPromise;
        if (getEnclavesResponseResult.isErr()) {
            return err(getEnclavesResponseResult.error);
        }

        return ok(getEnclavesResponseResult.value);
    }

    public async getUserServiceLogs(getUserServiceLogsArgs: GetUserServiceLogsArgs): Promise<Result<GetUserServiceLogsResponse, Error>> {
        const getUserServiceLogsPromise: Promise<Result<GetUserServiceLogsResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getUserServiceLogs(getUserServiceLogsArgs, (error: ServiceError | null, response?: GetUserServiceLogsResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)))
                    } else {
                        resolve(ok(response));
                    }
                } else {
                    resolve(err(error))
                }
            })
        })

        const getUserServiceLogsResponseResult: Result<GetUserServiceLogsResponse, Error> = await getUserServiceLogsPromise;
        if (getUserServiceLogsResponseResult.isErr()) {
            return err(getUserServiceLogsResponseResult.error);
        }

        return ok(getUserServiceLogsResponseResult.value);
    }

    public async streamUserServiceLogs(getUserServiceLogsArgs: GetUserServiceLogsArgs): Promise<Result<Readable, Error>> {

        const streamUserServiceLogsPromise: Promise<Result<ClientReadableStream<GetUserServiceLogsResponse>, Error>> = new Promise((resolve, _unusedReject) => {
            const getUserServiceLogsStreamResponse: ClientReadableStream<GetUserServiceLogsResponse> = this.client.streamUserServiceLogs(getUserServiceLogsArgs);
            resolve(ok(getUserServiceLogsStreamResponse));
        })

        const streamUserServiceLogsResponseResult: Result<ClientReadableStream<GetUserServiceLogsResponse>, Error> = await streamUserServiceLogsPromise;
        if (streamUserServiceLogsResponseResult.isErr()) {
            return err(streamUserServiceLogsResponseResult.error);
        }

        const streamUserServiceLogsResponse: ClientReadableStream<GetUserServiceLogsResponse> = streamUserServiceLogsResponseResult.value;

        const userServiceLogsByGuid: Map<ServiceGUID, Array<string>> = new Map<ServiceGUID, Array<string>>();

        const userServiceLogsReadable: Readable = this.createNewUserServiceLogsReadable(streamUserServiceLogsResponse);

        streamUserServiceLogsResponse.on('data', function (getUserServiceLogsResponse: GetUserServiceLogsResponse) {

            const userServiceLogsByUserServiceGuidMap: jspb.Map<string, LogLine> | undefined = getUserServiceLogsResponse.getUserServiceLogsByUserServiceGuidMap();

            if (userServiceLogsByUserServiceGuidMap !== undefined) {
                userServiceLogsByUserServiceGuidMap.forEach(
                    (userServiceLogLine, userServiceGuidStr) => {
                        userServiceLogsByGuid.set(userServiceGuidStr, userServiceLogLine.getLineList());
                    }
                )
            }

            userServiceLogsReadable.push(userServiceLogsByGuid);
        })

        streamUserServiceLogsResponse.on('error', (streamLogsErr) => {
            if (!userServiceLogsReadable.destroyed) {
                //Propagate the GRPC error to the user service logs readable
                const grpcStreamErr = new Error(`An error has been returned from the user service logs GRPC stream. Error:\n ${streamLogsErr}`)
                userServiceLogsReadable.emit('error', grpcStreamErr);
            }
        })

        streamUserServiceLogsResponse.on('end', function () {
            //Emit streams 'end' event when the GRPC stream has end
            if (!userServiceLogsReadable.destroyed) {
                userServiceLogsReadable.emit('end');
            }
        })

        return ok(userServiceLogsReadable);
    }

    private createNewUserServiceLogsReadable(streamUserServiceLogsResponse: ClientReadableStream<GetUserServiceLogsResponse>): Readable {
        const userServiceLogsReadable: Readable = new Readable({
            objectMode: true, //setting object mode is to allow pass objects in the readable.push() method
            read() {} //this is mandatory to implement, we implement empty as it's describe in the implementation examples here: https://nodesource.com/blog/understanding-streams-in-nodejs/
        })
        userServiceLogsReadable.on('close', function(){
            //Cancel the GRPC stream when when users close the ServiceLogsReadable
            streamUserServiceLogsResponse.cancel();
        })
        return userServiceLogsReadable;
    }
}
