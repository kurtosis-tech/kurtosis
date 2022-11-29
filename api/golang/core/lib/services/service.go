/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package services

/*
The identifier used for services within the enclave.
*/
type ServiceID string

/*
The globally unique identifier used for services within the enclave.
*/
type ServiceGUID string

type ServiceInfo struct {
	serviceId   ServiceID
	serviceGuid ServiceGUID
}

func NewServiceInfo(
	serviceId ServiceID,
	serviceGuid ServiceGUID,
) *ServiceInfo {
	return &ServiceInfo{
		serviceId:   serviceId,
		serviceGuid: serviceGuid,
	}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (serviceInfo *ServiceInfo) GetServiceID() ServiceID {
	return serviceInfo.serviceId
}

// Docs available at https://docs.kurtosistech.com/kurtosis-core/lib-documentation
func (serviceInfo *ServiceInfo) GetServiceGUID() ServiceGUID {
	return serviceInfo.serviceGuid
}
