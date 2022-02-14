import { Port } from "../../"

export type PortProtocol = Port.Protocol;
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
