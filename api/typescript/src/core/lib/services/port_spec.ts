import { Port } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export type TransportProtocol = Port.TransportProtocol;
export namespace TransportProtocol {
    export const TCP = Port.TransportProtocol.TCP;
    export const UDP = Port.TransportProtocol.UDP;
}

// Ports are 16 bit and should be no higher than max 16-bit number
export const MAX_PORT_NUM : number = 65535;

const   allowedTransportProtocols : Set<TransportProtocol> =
    new Set<TransportProtocol>([TransportProtocol.TCP, TransportProtocol.UDP]);

export function IsValidTransportProtocol(protocol: TransportProtocol): boolean {
    return allowedTransportProtocols.has(protocol)
}

export class PortSpec {
    constructor(
        public readonly number: number,
        public readonly transportProtocol: TransportProtocol,
        public readonly maybeApplicationProtocol?: string
    ) {}
    
    // No need for getters because the fields are 'readonly'
}
