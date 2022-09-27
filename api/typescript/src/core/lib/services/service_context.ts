import { err, ok, Result } from 'neverthrow';
import { newExecCommandArgs, newPauseServiceArgs, newUnpauseServiceArgs } from '../constructor_calls';
import type { ExecCommandArgs } from '../../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import type { PortSpec } from './port_spec';
import type { ServiceID } from './service';
import { GenericApiContainerClient } from '../enclaves/generic_api_container_client';
import {PauseServiceArgs, UnpauseServiceArgs} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ServiceContext {
    constructor(
        private readonly client: GenericApiContainerClient,
        private readonly serviceId: ServiceID,
        private readonly privateIpAddress: string,
        private readonly privatePorts: Map<string, PortSpec>,
        private readonly publicIpAddress: string,
        private readonly publicPorts: Map<string, PortSpec>,
    ) {}

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getServiceID(): ServiceID { 
        return this.serviceId;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivateIPAddress(): string {
        return this.privateIpAddress
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivatePorts(): Map<string, PortSpec> {
        return this.privatePorts
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getMaybePublicIPAddress(): string {
        return this.publicIpAddress
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPublicPorts(): Map<string, PortSpec> {
        return this.publicPorts
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execCommand(command: string[], ): Promise<Result<[number, string], Error>> {
        const execCommandArgs: ExecCommandArgs = newExecCommandArgs(this.serviceId, command);

        const execCommandResponseResult = await this.client.execCommand(execCommandArgs)
        if(execCommandResponseResult.isErr()){
            return err(execCommandResponseResult.error)
        }

        const execCommandResponse = execCommandResponseResult.value
        return ok([execCommandResponse.getExitCode(), execCommandResponse.getLogOutput()]);
    }
}
