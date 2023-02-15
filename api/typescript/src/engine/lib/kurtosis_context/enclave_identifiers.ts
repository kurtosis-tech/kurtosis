import {EnclaveUUID} from "../../../core/lib/enclaves/enclave_context";
import kurtosis_engine_rpc_api_bindings =  require("../../kurtosis_engine_rpc_api_bindings/engine_service_pb");

const VALID_UUID_MATCHES_ALLOWED = 1

// Docs available at https://docs.kurtosis.com/sdk#enclave-identifiers
export class EnclaveIdentifiers {
    public readonly enclaveNameToUuids: Map<string, EnclaveUUID[]>;
    public readonly enclaveUuids: Map<EnclaveUUID, boolean>;
    public readonly enclaveShortenedUuidToUuids: Map<string, EnclaveUUID[]>;

    constructor(historicalIdentifiers : kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers[]) {
        this.enclaveUuids = new Map<string, boolean>();
        this.enclaveNameToUuids = new Map<string, EnclaveUUID[]>();
        this.enclaveShortenedUuidToUuids = new Map<string, EnclaveUUID[]>();
        historicalIdentifiers.forEach((enclaveIdentifiers) => {
            let enclaveName = enclaveIdentifiers.getName();
            let enclaveUuid : EnclaveUUID = enclaveIdentifiers.getEnclaveUuid();
            let shortenedUuid = enclaveIdentifiers.getShortenedUuid();

            this.enclaveUuids.set(enclaveUuid, true);

            if (!(enclaveName in this.enclaveNameToUuids)) {
                this.enclaveNameToUuids.set(enclaveName, []);
            }
            let enclaveUuids = this.enclaveNameToUuids.get(enclaveName)!
            enclaveUuids.push(enclaveUuid)
            this.enclaveNameToUuids.set(enclaveName, enclaveUuids)

            if (!(shortenedUuid in this.enclaveShortenedUuidToUuids)) {
                this.enclaveShortenedUuidToUuids.set(shortenedUuid, []);
            }
            enclaveUuids = this.enclaveShortenedUuidToUuids.get(shortenedUuid)!
            enclaveUuids.push(enclaveUuid)
            this.enclaveShortenedUuidToUuids.set(shortenedUuid, enclaveUuids)
        });
    }

    public getEnclaveUuidForIdentifier(identifier: string): EnclaveUUID {
        if (this.enclaveUuids.has(identifier)) {
            return identifier as EnclaveUUID
        }

        if (this.enclaveShortenedUuidToUuids.has(identifier)) {
            let matches = this.enclaveShortenedUuidToUuids.get(identifier)!
            if (matches.length === VALID_UUID_MATCHES_ALLOWED) {
                return matches[0];
            } else if (matches.length > VALID_UUID_MATCHES_ALLOWED) {
                throw new Error(`Found multiple enclaves ${matches} matching shortened uuid ${identifier}. Please use a uuid to be more specific`)
            }
        }

        if (this.enclaveNameToUuids.has(identifier)) {
            let matches = this.enclaveNameToUuids.get(identifier)!
            if (matches.length === VALID_UUID_MATCHES_ALLOWED) {
                return matches[0];
            } else if (matches.length > VALID_UUID_MATCHES_ALLOWED) {
                throw new Error(`Found multiple enclaves ${matches} matching name ${identifier}. Please use a uuid to be more specific`)
            }
        }

        throw new Error(`No matching uuid for identifier ${identifier}`)
    }

    public getOrderedListOfNamesAndUuids(): String[] {
        let enclaveNames: string[] = [];
        let enclaveUuids: string[] = [];

        for (let name in this.enclaveNameToUuids) {
            enclaveNames.push(name);
        }

        for (let uuid in this.enclaveUuids) {
            enclaveUuids.push(uuid);
        }

        enclaveNames.sort();
        enclaveUuids.sort();

        return [...enclaveNames, ...enclaveUuids];
    }
}
