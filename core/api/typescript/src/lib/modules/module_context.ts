import { ApiContainerServiceClient } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { ExecuteModuleArgs, ExecuteModuleResponse } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { newExecuteModuleArgs as newExecuteModuleArgs } from "../constructor_calls";
import { ok, err, Result } from "neverthrow";
import * as grpc from "grpc";

export type ModuleID = string;

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
export class ModuleContext {
    private readonly client: ApiContainerServiceClient;
    private readonly moduleId: ModuleID;
    
    constructor (client: ApiContainerServiceClient, moduleId: ModuleID) {
        this.client = client;
        this.moduleId = moduleId;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
    public async execute(serializedParams: string): Promise<Result<string, Error>> {
        const args: ExecuteModuleArgs = newExecuteModuleArgs(this.moduleId, serializedParams);

        const executeModulePromise: Promise<Result<ExecuteModuleResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeModule(args, (error: grpc.ServiceError | null, response?: ExecuteModuleResponse) => {
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
        if (!executeModuleResult.isOk()) {
            return err(executeModuleResult.error);
        }
        const resp: ExecuteModuleResponse = executeModuleResult.value;

        return ok(resp.getSerializedResult());
    }
}
