import { Port } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb"

export type PortProtocol = Port.ProtocolMap[keyof Port.ProtocolMap];
export namespace PortProtocol {
    export const TCP = Port.Protocol.TCP;
    export const UDP = Port.Protocol.UDP;
}

export class PortSpec {
    constructor(
        public readonly number: number,
        public readonly protocol: PortProtocol,
    ) {}
    
    // No need for getters because the fields are 'readonly'
}