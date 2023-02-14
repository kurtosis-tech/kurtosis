import {EnclaveUUID} from "../../../core/lib/enclaves/enclave_context";
import {EnclaveInfo} from "../../kurtosis_engine_rpc_api_bindings/engine_service_pb";

// Enclaves A collection of enclaves by uuid, name and shortened uuid
export class Enclaves {
    constructor(
        public readonly enclavesByUuid: Map<EnclaveUUID, EnclaveInfo>,
        public readonly enclavesByName: Map<string, EnclaveInfo>,
        public readonly enclavesByShortenedUuid: Map<string, EnclaveInfo[]>
    ){}

    public getEnclavesByUuid(): Map<EnclaveUUID, EnclaveInfo> {
        return this.enclavesByUuid
    }

    public getEnclavesByName(): Map<string, EnclaveInfo> {
        return this.enclavesByName
    }

    public getEnclavesByShortenedUuid(): Map<string, EnclaveInfo[]> {
        return this.enclavesByShortenedUuid
    }
}
