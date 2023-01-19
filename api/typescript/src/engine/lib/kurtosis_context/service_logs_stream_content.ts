import {ServiceUUID} from "../../../core/lib/services/service";
import {ServiceLog} from "./service_log";

//This struct wrap the information returned by the user service logs GRPC stream
export class ServiceLogsStreamContent {
    private readonly serviceLogsByServiceUuids: Map<ServiceUUID, Array<ServiceLog>>;
    private readonly notFoundServiceUuids: Set<ServiceUUID>;

    constructor(
        serviceLogsByServiceUuids: Map<ServiceUUID, Array<ServiceLog>>,
        notFoundServiceUuids: Set<ServiceUUID>,
    ) {
        this.serviceLogsByServiceUuids = serviceLogsByServiceUuids;
        this.notFoundServiceUuids = notFoundServiceUuids;
    }

    // Docs available at https://docs.kurtosis.com/sdk#getservicelogsbyserviceuuids----mapserviceuuid-arrayservicelog-servicelogsbyserviceuuids
    public getServiceLogsByServiceUuids(): Map<ServiceUUID, Array<ServiceLog>> {
        return this.serviceLogsByServiceUuids;
    }

    // Docs available at https://docs.kurtosis.com/sdk#getnotfoundserviceuuids---setserviceuuid-notfoundserviceuuids
    public getNotFoundServiceUuids(): Set<ServiceUUID> {
        return this.notFoundServiceUuids;
    }
}
