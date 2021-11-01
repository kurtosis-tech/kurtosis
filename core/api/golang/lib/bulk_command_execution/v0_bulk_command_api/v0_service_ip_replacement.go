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

package v0_bulk_command_api

const (
	ServiceIdIpReplacementPrefix = "<<<"
	ServiceIdIpReplacementSuffix = ">>>"
)

// Used to encode a service ID to a string that can be embedded in commands, and which the API container will replace
//  with the IP address of the service at runtime
func EncodeServiceIdForIpReplacement(serviceId string) string {
	return ServiceIdIpReplacementPrefix + serviceId + ServiceIdIpReplacementSuffix
}
