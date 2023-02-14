import {ServiceName, ServiceUUID} from "./service";
import kurtosis_core_rpc_api_bindings =  require("../../kurtosis_core_rpc_api_bindings/api_container_service_pb")

const VALID_UUID_MATCHES_ALLOWED = 1

// Docs available at https://docs.kurtosis.com/sdk#service-identifiers
export class ServiceIdentifiers {
    public readonly serviceNameToUuids: Map<ServiceName, ServiceUUID[]>;
    public readonly serviceUuids: Map<ServiceUUID, boolean>;
    public readonly serviceShortenedUuidToUuids: Map<string, ServiceUUID[]>;

    constructor(historicalIdentifiers : kurtosis_core_rpc_api_bindings.ServiceIdentifiers[]) {
        this.serviceUuids = new Map<string, boolean>();
        this.serviceNameToUuids = new Map<ServiceName, ServiceUUID[]>();
        this.serviceShortenedUuidToUuids = new Map<string, ServiceUUID[]>();
        historicalIdentifiers.forEach((serviceIdentifiers) => {
            let serviceName = serviceIdentifiers.getName();
            let serviceUuid : ServiceUUID = serviceIdentifiers.getServiceUuid();
            let shortenedUuid = serviceIdentifiers.getShortenedUuid();

            this.serviceUuids.set(serviceUuid, true);

            if (!(serviceName in this.serviceNameToUuids)) {
                this.serviceNameToUuids.set(serviceName, []);
            }
            let serviceUuids = this.serviceNameToUuids.get(serviceName)!
            serviceUuids.push(serviceUuid)
            this.serviceNameToUuids.set(serviceName, serviceUuids)

            if (!(shortenedUuid in this.serviceShortenedUuidToUuids)) {
                this.serviceShortenedUuidToUuids.set(shortenedUuid, []);
            }
            serviceUuids = this.serviceShortenedUuidToUuids.get(shortenedUuid)!
            serviceUuids.push(serviceUuid)
            this.serviceShortenedUuidToUuids.set(shortenedUuid, serviceUuids)
        });
    }

    public getServiceUuidForIdentifier(identifier: string): ServiceUUID {
        if (this.serviceUuids.has(identifier)) {
            return identifier as ServiceUUID
        }

        if (this.serviceShortenedUuidToUuids.has(identifier)) {
            let matches = this.serviceShortenedUuidToUuids.get(identifier)!
            if (matches.length === VALID_UUID_MATCHES_ALLOWED) {
                return matches[0];
            } else if (matches.length > VALID_UUID_MATCHES_ALLOWED) {
                throw new Error(`Found multiple services ${matches} matching shortened uuid ${identifier}. Please use a uuid to be more specific`)
            }
        }

        if (this.serviceNameToUuids.has(identifier)) {
            let matches = this.serviceNameToUuids.get(identifier)!
            if (matches.length === VALID_UUID_MATCHES_ALLOWED) {
                return matches[0];
            } else if (matches.length > VALID_UUID_MATCHES_ALLOWED) {
                throw new Error(`Found multiple services ${matches} matching name ${identifier}. Please use a uuid to be more specific`)
            }
        }

        throw new Error(`No matching uuid for identifier ${identifier}`)
    }

    public getOrderedListOfNamesAndUuids(): String[] {
        let serviceNames: string[] = [];
        let serviceUuids: string[] = [];

        for (let name in this.serviceNameToUuids) {
            serviceNames.push(name);
        }

        for (let uuid in this.serviceUuids) {
            serviceUuids.push(uuid);
        }

        serviceNames.sort();
        serviceUuids.sort();

        return [...serviceNames, ...serviceUuids];
    }
}
