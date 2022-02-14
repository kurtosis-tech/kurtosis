import { Result } from 'neverthrow';
import { 
    ApiContainerServiceClientNode, 
    ApiContainerServiceClientWeb,
    ServiceID,
    SharedPath,
    PortSpec 
} from '../..';
import { GrpcNodeServiceContextBackend } from './service_context_backend_node';
import { GrpcWebServiceContextBackend } from './service_context_backend_web';

export interface ServiceContextBackend {
    execCommand(command: string[], serviceId: ServiceID): Promise<Result<[number, string], Error>>
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
export class ServiceContext {

    private backend: ServiceContextBackend
    private readonly serviceId: ServiceID
    private readonly sharedDirectory: SharedPath
    private readonly privateIpAddr: string
    private readonly privatePorts: Map<string, PortSpec>
    private readonly publicIpAddr: string
    private readonly publicPorts: Map<string, PortSpec>

    constructor(
        client: ApiContainerServiceClientWeb | ApiContainerServiceClientNode,
        serviceId: ServiceID,
        sharedDirectory: SharedPath,
        privateIpAddr: string,
        privatePorts: Map<string, PortSpec>,
        publicIpAddr: string,
        publicPorts: Map<string, PortSpec>
        ){

        if(client instanceof ApiContainerServiceClientWeb){
            this.backend = new GrpcWebServiceContextBackend(client)
        }else{
            this.backend = new GrpcNodeServiceContextBackend(client)
        }

        this.serviceId = serviceId
        this.sharedDirectory = sharedDirectory
        this.privateIpAddr = privateIpAddr
        this.privatePorts = privatePorts
        this.publicIpAddr = publicIpAddr
        this.publicPorts = publicPorts
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getServiceID(): ServiceID { 
        return this.getServiceID();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getSharedDirectory(): SharedPath {
        return  this.getSharedDirectory()
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivateIPAddress(): string {
        return this.getPrivateIPAddress();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPrivatePorts(): Map<string, PortSpec> {
        return this.getPrivatePorts();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getMaybePublicIPAddress(): string {
        return this.getMaybePublicIPAddress();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public getPublicPorts(): Map<string, PortSpec> {
        return this.getPublicPorts();
    }

    // Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
    public async execCommand(command: string[], ): Promise<Result<[number, string], Error>> {
       return this.backend.execCommand(command, this.serviceId)
    }
}
