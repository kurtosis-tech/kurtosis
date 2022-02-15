import { ok, err, Result } from "neverthrow";
import * as grpc_web from "grpc-web";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ExecuteModuleArgs, ExecuteModuleResponse } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { ModuleContextBackend } from "./module_context_backend";

export class GrpcWebModuleContextBackend implements ModuleContextBackend{
    private readonly client: ApiContainerServiceClientWeb;
    
    constructor (client: ApiContainerServiceClientWeb) {
        this.client = client;
    }

    public async execute(executeModuleArgs: ExecuteModuleArgs): Promise<Result<ExecuteModuleResponse, Error>> {
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
}