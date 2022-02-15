import { Result } from "neverthrow";
import { ExecuteModuleArgs } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export interface ModuleContextBackend {
    execute(executeModuleArgs: ExecuteModuleArgs): Promise<Result<string, Error>>
}