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
        private readonly serviceID: ServiceID,
        private readonly serviceGUID: ServiceGUID,
    ) {
    }

    public getServiceID(): ServiceID {
        return this.serviceID;
    }

    public getServiceGUID(): ServiceGUID {
        return this.serviceGUID;
    }
}