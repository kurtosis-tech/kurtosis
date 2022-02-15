import * as grpc_web from "grpc-web";
import { ok, err, Result } from 'neverthrow';
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ExecCommandArgs, ExecCommandResponse } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ServiceContextBackend } from "./service_context_interface";

export class GrpcWebServiceContextBackend implements ServiceContextBackend {
    private readonly client: ApiContainerServiceClientWeb
    constructor(client: ApiContainerServiceClientWeb) {
        this.client = client
    }

    public async execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>> {
        const promiseExecCommand: Promise<Result<ExecCommandResponse, Error>> = new Promise((resolve, _unusedReject) => {
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
        const resultExecCommand: Result<ExecCommandResponse, Error> = await promiseExecCommand;
        
        return resultExecCommand
    }
}