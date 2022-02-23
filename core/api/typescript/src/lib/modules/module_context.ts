import { err, ok, Result } from "neverthrow";
import { ExecuteModuleArgs } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { newExecuteModuleArgs } from "../constructor_calls";
import { GenericApiContainerClient } from "../enclaves/generic_api_container_client";

export type ModuleID= string;

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ModuleContext {

    private readonly client: GenericApiContainerClient
    private readonly moduleId: ModuleID;

    constructor (client: GenericApiContainerClient, moduleId: ModuleID) {
        this.moduleId = moduleId
        this.client = client
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execute(serializedParams: string): Promise<Result<string, Error>> {
        const executeModuleArgs: ExecuteModuleArgs = newExecuteModuleArgs(this.moduleId, serializedParams);

        const executeResponseResult = await this.client.executeModule(executeModuleArgs)
        if(executeResponseResult.isErr()){
            return err(executeResponseResult.error)
        }

        const executeResponse = executeResponseResult.value
        return ok(executeResponse.getSerializedResult())
    }
}
