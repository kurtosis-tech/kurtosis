package enclave_structure

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type EnclaveComponents struct {
	enclaveServices      map[service.ServiceName]EnclaveComponentStatus
	enclaveFilesArtifact map[string]EnclaveComponentStatus
}

func NewEnclaveComponents() *EnclaveComponents {
	return &EnclaveComponents{
		enclaveServices:      map[service.ServiceName]EnclaveComponentStatus{},
		enclaveFilesArtifact: map[string]EnclaveComponentStatus{},
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

func (components *EnclaveComponents) AddFilesArtifact(artifactName string, componentStatus EnclaveComponentStatus) {
	components.enclaveFilesArtifact[artifactName] = componentStatus
}

func (components *EnclaveComponents) HasFilesArtifactBeenUpdated(artifactName string) bool {
	if serviceStatus, found := components.enclaveFilesArtifact[artifactName]; found {
		return serviceStatus == ComponentIsUpdated
	}
	return false
}
