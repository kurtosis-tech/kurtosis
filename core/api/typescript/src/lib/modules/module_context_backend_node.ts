import { ok, err, Result } from "neverthrow";
import * as grpc_node from "@grpc/grpc-js";
import { 
    ApiContainerServiceClientNode, 
    newExecuteModuleArgs,
    ExecuteModuleArgs, 
    ExecuteModuleResponse
} from "../..";
import { ModuleContextBackend, ModuleID } from "./module_context";

export class GrpcNodeModuleContextBackend implements ModuleContextBackend{
    private readonly client: ApiContainerServiceClientNode;
    
    constructor (client: ApiContainerServiceClientNode) {
        this.client = client;
    }

    public async execute(serializedParams: string, moduleId: ModuleID): Promise<Result<string, Error>> {
        const args: ExecuteModuleArgs = newExecuteModuleArgs(moduleId, serializedParams);

        const executeModulePromise: Promise<Result<ExecuteModuleResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeModule(args, (error: grpc_node.ServiceError | null, response?: ExecuteModuleResponse) => {
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
        const executeModuleResponse: ExecuteModuleResponse = executeModuleResult.value;

        return ok(executeModuleResponse.getSerializedResult());
    }
}