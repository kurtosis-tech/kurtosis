import {ServiceGUID} from "../../../core/lib/services/service";
import {ServiceLog} from "./service_log";

//This struct wrap the information returned by the user service logs GRPC stream
export class ServiceLogsStreamContent {
    private readonly serviceLogsByServiceGuids: Map<ServiceGUID, Array<ServiceLog>>;
    private readonly notFoundServiceGuids: Set<ServiceGUID>;

    constructor(
        serviceLogsByServiceGuids: Map<ServiceGUID, Array<ServiceLog>>,
        notFoundServiceGuids: Set<ServiceGUID>,
    ) {
        this.serviceLogsByServiceGuids = serviceLogsByServiceGuids;
        this.notFoundServiceGuids = notFoundServiceGuids;
    }

    public getServiceLogsByServiceGuids(): Map<ServiceGUID, Array<ServiceLog>> {
        return this.serviceLogsByServiceGuids;
    }

    public getNotFoundServiceGuids(): Set<ServiceGUID> {
        return this.notFoundServiceGuids;
    }
}
