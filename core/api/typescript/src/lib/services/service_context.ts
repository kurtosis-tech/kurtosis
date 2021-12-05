import { ApiContainerServiceClient } from '../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb'; 
import { ExecCommandArgs, ExecCommandResponse } from '../../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { ServiceID} from './service';
import { SharedPath } from './shared_path';
import { newExecCommandArgs} from "../constructor_calls";
import { ok, err, Result } from 'neverthrow';
import * as grpc from "@grpc/grpc-js";
import { PortSpec } from './port_spec';

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ServiceContext {
    
    constructor(
        private readonly client: ApiContainerServiceClient,
        private readonly serviceId: ServiceID,
        private readonly sharedDirectory: SharedPath,
        private readonly privateIpAddr: string,
        private readonly privatePorts: Map<string, PortSpec>,
        private readonly publicIpAddr: string,
        private readonly publicPorts: Map<string, PortSpec>,
    ) {}

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getServiceID(): ServiceID { 
        return this.serviceId;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getSharedDirectory(): SharedPath {
        return  this.sharedDirectory
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivateIPAddress(): string {
        return this.privateIpAddr;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivatePorts(): Map<string, PortSpec> {
        return this.privatePorts;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getMaybePublicIPAddress(): string {
        return this.publicIpAddr;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPublicPorts(): Map<string, PortSpec> {
        return this.publicPorts;
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execCommand(command: string[]): Promise<Result<[number, string], Error>> {
        const serviceId: ServiceID = this.serviceId;
        const args: ExecCommandArgs = newExecCommandArgs(serviceId, command);

        const promiseExecCommand: Promise<Result<ExecCommandResponse, Error>> = new Promise((resolve, _unusedReject) => {
            this.client.execCommand(args, (error: grpc.ServiceError | null, response?: ExecCommandResponse) => {
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
        if (!resultExecCommand.isOk()) {
            return err(resultExecCommand.error);
        }
        const resp: ExecCommandResponse = resultExecCommand.value;

        return ok([resp.getExitCode(), resp.getLogOutput()]);
    }
}
