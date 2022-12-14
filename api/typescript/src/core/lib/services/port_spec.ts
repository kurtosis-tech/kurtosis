import { Port } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export type PortProtocol = Port.TransportProtocol;
export namespace PortProtocol {
    export const TCP = Port.TransportProtocol.TCP;
    export const UDP = Port.TransportProtocol.UDP;
}

export class PortSpec {
    constructor(
        public readonly number: number,
        public readonly protocol: PortProtocol,
    ) {}
    
    // No need for getters because the fields are 'readonly'
}
