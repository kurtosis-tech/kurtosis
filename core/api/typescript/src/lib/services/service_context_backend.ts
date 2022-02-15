import { Result } from "neverthrow";
import { ExecCommandArgs, ExecCommandResponse } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export interface ServiceContextBackend {
    execCommand(execCommandArgs: ExecCommandArgs): Promise<Result<ExecCommandResponse, Error>>
}