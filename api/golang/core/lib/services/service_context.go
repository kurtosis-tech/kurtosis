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

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/stacktrace"
)

type ServiceContext struct {
	client      kurtosis_core_rpc_api_bindings.ApiContainerServiceClient
	serviceName ServiceName
	serviceUuid ServiceUUID

	// Network location inside the enclave
	privateIpAddr string
	privatePorts  map[string]*PortSpec

	// Network location outside the enclave
	publicIpAddr string
	publicPorts  map[string]*PortSpec

	// Service labels
	labels map[string]string
}

func NewServiceContext(
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient,
	serviceName ServiceName,
	serviceUuid ServiceUUID,
	privateIpAddr string,
	privatePorts map[string]*PortSpec,
	publicIpAddr string,
	publicPorts map[string]*PortSpec,
	labels map[string]string,
) *ServiceContext {
	return &ServiceContext{
		client:        client,
		serviceName:   serviceName,
		serviceUuid:   serviceUuid,
		privateIpAddr: privateIpAddr,
		privatePorts:  privatePorts,
		publicIpAddr:  publicIpAddr,
		publicPorts:   publicPorts,
		labels:        labels,
	}
}

func (service *ServiceContext) GetServiceName() ServiceName {
	return service.serviceName
}

func (service *ServiceContext) GetServiceUUID() ServiceUUID {
	return service.serviceUuid
}

func (service *ServiceContext) GetPrivateIPAddress() string {
	return service.privateIpAddr
}

func (service *ServiceContext) GetPrivatePorts() map[string]*PortSpec {
	return service.privatePorts
}

func (service *ServiceContext) GetMaybePublicIPAddress() string {
	return service.publicIpAddr
}

func (service *ServiceContext) GetPublicPorts() map[string]*PortSpec {
	return service.publicPorts
}

func (service *ServiceContext) GetLabels() map[string]string {
	return service.labels
}

func (service *ServiceContext) ExecCommand(command []string) (int32, string, error) {
	serviceName := service.serviceName
	args := binding_constructors.NewExecCommandArgs(string(serviceName), command)
	resp, err := service.client.ExecCommand(context.Background(), args)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred executing command '%v' on service '%v'",
			command,
			serviceName)
	}
	return resp.ExitCode, resp.LogOutput, nil
}
