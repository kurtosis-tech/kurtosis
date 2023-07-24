package enclave_structure

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type EnclaveComponents struct {
	enclaveServices map[service.ServiceName]EnclaveComponentStatus
}

func NewEnclaveComponents() *EnclaveComponents {
	return &EnclaveComponents{
		enclaveServices: map[service.ServiceName]EnclaveComponentStatus{},
	}
}

func (components *EnclaveComponents) AddService(serviceName service.ServiceName, componentStatus EnclaveComponentStatus) {
	components.enclaveServices[serviceName] = componentStatus
}

func (components *EnclaveComponents) HasServiceBeenUpdated(serviceName service.ServiceName) bool {
	if serviceStatus, found := components.enclaveServices[serviceName]; found {
		return serviceStatus == ComponentIsUpdated
	}
	return false
}
