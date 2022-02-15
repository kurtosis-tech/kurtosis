import { err, ok, Result } from "neverthrow";
import { newExecuteModuleArgs } from "../constructor_calls";
import { GrpcNodeModuleContextBackend } from "./grpc_node_module_context_backend";
import { GrpcWebModuleContextBackend } from "./grpc_web_module_context_backend";
import { ModuleContextBackend } from "./module_context_backend";
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientNode } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { ExecuteModuleArgs } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export type ModuleID= string;

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ModuleContext {
    private readonly backend: ModuleContextBackend
    private readonly moduleId: ModuleID;

    
    constructor (client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode, moduleId: ModuleID) {
        if(client instanceof ApiContainerServiceClientWeb){
            this.backend = new GrpcWebModuleContextBackend(client)
        }else{
            this.backend = new GrpcNodeModuleContextBackend(client)
        }

        this.moduleId = moduleId
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execute(serializedParams: string): Promise<Result<string, Error>> {
        const executeModuleArgs: ExecuteModuleArgs = newExecuteModuleArgs(this.moduleId, serializedParams);

        const executeResponseResult = await this.backend.execute(executeModuleArgs)
        if(executeResponseResult.isErr()){
            return err(executeResponseResult.error)
        }

        const executeResponse = executeResponseResult.value
        return ok(executeResponse.getSerializedResult())
    }
}
