import { Port } from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";

export type TransportProtocol = Port.TransportProtocol;
export namespace TransportProtocol {
    export const TCP = Port.TransportProtocol.TCP;
    export const UDP = Port.TransportProtocol.UDP;
}

export class PortSpec {
    constructor(
        public readonly number: number,
        public readonly transportProtocol: TransportProtocol,
        public readonly maybeApplicationProtocol?: string
    ) {}
    
    // No need for getters because the fields are 'readonly'
}
