import { err, ok, Result } from 'neverthrow';
import { ExecCommandArgs } from '../../kurtosis_core_rpc_api_bindings/api_container_service_pb';
import { newExecCommandArgs } from '../constructor_calls';
import { ApiContainerServiceClient as ApiContainerServiceClientWeb } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_web_pb";
import { ApiContainerServiceClient as ApiContainerServiceClientNode } from "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb";
import { GrpcNodeServiceContextBackend } from './grpc_node_service_context_backend';
import { GrpcWebServiceContextBackend } from './grpc_web_service_context_backend';
import { PortSpec } from './port_spec';
import { ServiceID } from './service';
import { ServiceContextBackend } from './service_context_backend';
import { SharedPath } from './shared_path';

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ServiceContext {

    private readonly backend: ServiceContextBackend
    private readonly serviceId: ServiceID
    private readonly sharedDirectory: SharedPath
    private readonly privateIpAddress: string
    private readonly privatePorts: Map<string, PortSpec>
    private readonly publicIpAddress: string
    private readonly publicPorts: Map<string, PortSpec>

    constructor(
        client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode,
        serviceId: ServiceID,
        sharedDirectory: SharedPath,
        privateIpAddress: string,
        privatePorts: Map<string, PortSpec>,
        publicIpAddress: string,
        publicPorts: Map<string, PortSpec>
        ){

        if(client instanceof ApiContainerServiceClientWeb){
            this.backend = new GrpcWebServiceContextBackend(client)
        }else{
            this.backend = new GrpcNodeServiceContextBackend(client)
        }

        this.serviceId = serviceId
        this.sharedDirectory = sharedDirectory
        this.privateIpAddress = privateIpAddress
        this.privatePorts = privatePorts
        this.publicIpAddress = publicIpAddress
        this.publicPorts = publicPorts
    }

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

        const execCommandResponseResult = await this.backend.execCommand(execCommandArgs)
        if(execCommandResponseResult.isErr()){
            return err(execCommandResponseResult.error)
        }

        const execCommandResponse = execCommandResponseResult.value

        return ok([execCommandResponse.getExitCode(), execCommandResponse.getLogOutput()]);
    }
}
