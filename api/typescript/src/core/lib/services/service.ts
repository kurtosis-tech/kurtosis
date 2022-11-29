/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

/*
The identifier used for services within the enclave.
*/

export type ServiceID = string;

/*
The globally unique identifier used for services within the enclave.
*/
export type ServiceGUID = string;

export class ServiceInfo {
    constructor(
        private readonly serviceId: ServiceID,
        private readonly serviceGuid: ServiceGUID,
    ) {
    }

    public getServiceId(): ServiceID {
        return this.serviceId;
    }

    public getServiceGuid(): ServiceGUID {
        return this.serviceGuid;
    }
}
