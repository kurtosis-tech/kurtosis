import { ok, err, Result } from "neverthrow";
import * as grpc_web from "grpc-web";
import { 
    ApiContainerServiceClientWeb,
    newExecuteModuleArgs,
    ExecuteModuleArgs, 
    ExecuteModuleResponse
} from "../..";
import { ModuleContextBackend, ModuleID } from "./module_context";



export class GrpcWebModuleContextBackend implements ModuleContextBackend{
    private readonly client: ApiContainerServiceClientWeb;
    
    constructor (client: ApiContainerServiceClientWeb) {
        this.client = client;
    }

    public async execute(serializedParams: string, moduleId: ModuleID): Promise<Result<string, Error>> {
        const args: ExecuteModuleArgs = newExecuteModuleArgs(moduleId, serializedParams);

        const executeModulePromise: Promise<Result<ExecuteModuleResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.executeModule(args, {}, (error: grpc_web.RpcError | null, response?: ExecuteModuleResponse) => {
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