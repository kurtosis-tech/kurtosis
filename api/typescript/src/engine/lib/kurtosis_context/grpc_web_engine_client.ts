import * as grpc_web from "grpc-web";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import {err, ok, Result} from "neverthrow";
import type { EngineServiceClient as EngineServiceClientWeb } from "../../kurtosis_engine_rpc_api_bindings/engine_service_grpc_web_pb";
import type { GenericEngineClient } from "./generic_engine_client";
import type {
    CleanArgs,
    CleanResponse,
    CreateEnclaveArgs,
    CreateEnclaveResponse,
    DestroyEnclaveArgs,
    GetEnclavesResponse,
    GetEngineInfoResponse,
    StopEnclaveArgs,
    GetServiceLogsArgs,
    GetServiceLogsResponse
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import {NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG} from "../consts";
import * as jspb from "google-protobuf";
import {
    GetExistingAndHistoricalEnclaveIdentifiersResponse,
    LogLine
} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";
import {ServiceUUID} from "../../../core/lib/services/service";
import {Readable} from "stream";
import {ServiceLog} from "./service_log";
import {ServiceLogsStreamContent} from "./service_logs_stream_content";

const GRPC_WEB_STREAM_RESPONSE_DATA_EVENT_NAME = 'data'
const GRPC_WEB_STREAM_RESPONSE_ERROR_EVENT_NAME = 'error'
const GRPC_WEB_STREAM_RESPONSE_END_EVENT_NAME = 'end'

const SERVICE_LOGS_READABLE_ERROR_EVENT_NAME = 'error'
const SERVICE_LOGS_READABLE_END_EVENT_NAME = 'end'
const SERVICE_LOGS_READABLE_CLOSE_EVENT_NAME = 'close'

export class GrpcWebEngineClient implements GenericEngineClient {
    private readonly client: EngineServiceClientWeb

    constructor(client: EngineServiceClientWeb){
        this.client = client
    }

    public async getEngineInfo(): Promise<Result<GetEngineInfoResponse,Error>> {
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()

        const getEngineInfoPromise: Promise<Result<GetEngineInfoResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEngineInfo(emptyArg, {}, (error: grpc_web.RpcError | null, response?: GetEngineInfoResponse) => {
                if (error === null) {
                    if (!response) {
                        resolve(err(new Error(NO_ERROR_ENCOUNTERED_BUT_RESPONSE_FALSY_MSG)));
                    } else {
                        resolve(ok(response!));
                    }
                } else {
                    if(error.code === grpc_web.StatusCode.UNAVAILABLE){
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

    public async createEnclaveResponse(createEnclaveArgs: CreateEnclaveArgs): Promise<Result<CreateEnclaveResponse, Error>> {
        const createEnclavePromise: Promise<Result<CreateEnclaveResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.createEnclave(createEnclaveArgs, {}, (error: grpc_web.RpcError | null, response?: CreateEnclaveResponse) => {
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
            this.client.stopEnclave(stopEnclaveArgs, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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
            this.client.destroyEnclave(destroyEnclaveArgs, {}, (error: grpc_web.RpcError | null, _unusedResponse?: google_protobuf_empty_pb.Empty) => {
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

    public async clean(cleanArgs: CleanArgs): Promise<Result<CleanResponse, Error>>{
        const cleanPromise: Promise<Result<CleanResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.clean(cleanArgs, {}, (error: grpc_web.RpcError | null, response?: CleanResponse) => {
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
            return err(cleanResult.error)
        }

        const cleanResponse: CleanResponse = cleanResult.value;
        return ok(cleanResponse);
    }

    public async getEnclavesResponse(): Promise<Result<GetEnclavesResponse, Error>>{
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        const getEnclavesPromise: Promise<Result<GetEnclavesResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getEnclaves(emptyArg, {}, (error: grpc_web.RpcError | null, response?: GetEnclavesResponse) => {
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
            return err(getEnclavesResponseResult.error)
        }

        return ok(getEnclavesResponseResult.value);
    }

    public async getServiceLogs(getServiceLogsArgs: GetServiceLogsArgs): Promise<Result<Readable, Error>> {

        const streamServiceLogsPromise: Promise<Result<grpc_web.ClientReadableStream<GetServiceLogsResponse>, Error>> = new Promise((resolve, _unusedReject) => {
            const getServiceLogsStreamResponse: grpc_web.ClientReadableStream<GetServiceLogsResponse> = this.client.getServiceLogs(getServiceLogsArgs);
            resolve(ok(getServiceLogsStreamResponse));
        })

        const streamServiceLogsResponseResult: Result<grpc_web.ClientReadableStream<GetServiceLogsResponse>, Error> = await streamServiceLogsPromise;
        if (streamServiceLogsResponseResult.isErr()) {
            return err(streamServiceLogsResponseResult.error);
        }

        const streamServiceLogsResponse: grpc_web.ClientReadableStream<GetServiceLogsResponse> = streamServiceLogsResponseResult.value;

        const serviceLogsByUuid: Map<ServiceUUID, Array<ServiceLog>> = new Map<ServiceUUID, Array<ServiceLog>>();

        const serviceLogsReadable: Readable = this.createNewServiceLogsReadable(streamServiceLogsResponse);

        streamServiceLogsResponse.on(GRPC_WEB_STREAM_RESPONSE_DATA_EVENT_NAME, function (getServiceLogsResponse: GetServiceLogsResponse) {
            const serviceLogsByUserServiceUuidMap: jspb.Map<string, LogLine> | undefined = getServiceLogsResponse.getServiceLogsByServiceUuidMap();

            if (serviceLogsByUserServiceUuidMap !== undefined) {
                serviceLogsByUserServiceUuidMap.forEach(
                    (serviceLogLine, serviceUuidStr) => {
                        const serviceLogs: Array<ServiceLog> = Array<ServiceLog>();

                        serviceLogLine.getLineList().forEach((logLine:string) => {
                            const serviceLog: ServiceLog = new ServiceLog(logLine)
                            serviceLogs.push(serviceLog)
                        })

                        serviceLogsByUuid.set(serviceUuidStr, serviceLogs);
                    }
                )
            }

            const notFoundServiceUuidsMap: jspb.Map<string, boolean> = getServiceLogsResponse.getNotFoundServiceUuidSetMap()

            const notFoundServiceUuids: Set<ServiceUUID> = new Set<ServiceUUID>()

            notFoundServiceUuidsMap.forEach((isUuidInMap: boolean, serviceUuidStr: string) => {
                notFoundServiceUuids.add(serviceUuidStr)
            })

            const serviceLogsStreamContent: ServiceLogsStreamContent = new ServiceLogsStreamContent(serviceLogsByUuid, notFoundServiceUuids)

            serviceLogsReadable.push(serviceLogsStreamContent);
        })

        streamServiceLogsResponse.on(GRPC_WEB_STREAM_RESPONSE_ERROR_EVENT_NAME, (streamLogsErr: any) => {
            if (!serviceLogsReadable.destroyed) {
                //Propagate the GRPC error to the user service logs readable
                const grpcStreamErr = new Error(`An error has been returned from the service logs GRPC stream. Error:\n ${streamLogsErr}`)
                serviceLogsReadable.emit(SERVICE_LOGS_READABLE_ERROR_EVENT_NAME, grpcStreamErr);
            }
        })

        streamServiceLogsResponse.on(GRPC_WEB_STREAM_RESPONSE_END_EVENT_NAME, function () {
            //Emit streams 'end' event when the GRPC stream has end
            if (!serviceLogsReadable.destroyed) {
                serviceLogsReadable.emit(SERVICE_LOGS_READABLE_END_EVENT_NAME);
            }
        })

        return ok(serviceLogsReadable);
    }

    public async getExistingAndHistoricalEnclaveIdentifiers(): Promise<Result<GetExistingAndHistoricalEnclaveIdentifiersResponse, Error>>{
        const emptyArg: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty()
        const getExistingAndHistoricalEnclaveIdentifiersPromise: Promise<Result<GetExistingAndHistoricalEnclaveIdentifiersResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.getExistingAndHistoricalEnclaveIdentifiers(emptyArg, {}, (error: grpc_web.RpcError | null, response?: GetExistingAndHistoricalEnclaveIdentifiersResponse) => {
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

        const getExistingAndHistoricalEnclaveIdentifiersResult: Result<GetExistingAndHistoricalEnclaveIdentifiersResponse, Error> = await getExistingAndHistoricalEnclaveIdentifiersPromise;
        if (getExistingAndHistoricalEnclaveIdentifiersResult.isErr()) {
            return err(getExistingAndHistoricalEnclaveIdentifiersResult.error)
        }

        return ok(getExistingAndHistoricalEnclaveIdentifiersResult.value);
    }

    private createNewServiceLogsReadable(streamServiceLogsResponse: grpc_web.ClientReadableStream<GetServiceLogsResponse>): Readable {
        const serviceLogsReadable: Readable = new Readable({
            objectMode: true, //setting object mode is to allow pass objects in the readable.push() method
            read() {} //this is mandatory to implement, we implement empty as it's describe in the implementation examples here: https://nodesource.com/blog/understanding-streams-in-nodejs/
        })
        serviceLogsReadable.on(SERVICE_LOGS_READABLE_CLOSE_EVENT_NAME, function(){
            //Cancel the GRPC stream when when users close the ServiceLogsReadable
            streamServiceLogsResponse.cancel();
        })
        return serviceLogsReadable;
    }
}
