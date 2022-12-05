package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	requiredDockerImages map[string]bool
	serviceIDs           map[service.ServiceID]bool
	artifactIDs          map[enclave_data_directory.FilesArtifactID]bool
}

func NewValidatorEnvironment(serviceIDs map[service.ServiceID]bool) *ValidatorEnvironment {
	return &ValidatorEnvironment{
		requiredDockerImages: map[string]bool{},
		serviceIDs:           serviceIDs,
		artifactIDs:          map[enclave_data_directory.FilesArtifactID]bool{},
	}
}

func (environment *ValidatorEnvironment) AppendRequiredDockerImage(dockerImage string) {
	environment.requiredDockerImages[dockerImage] = true
}

func (environment *ValidatorEnvironment) AddServiceId(serviceId service.ServiceID) {
	environment.serviceIDs[serviceId] = true
}

func (environment *ValidatorEnvironment) RemoveServiceId(serviceId service.ServiceID) {
	delete(environment.serviceIDs, serviceId)
}

func (environment *ValidatorEnvironment) DoesServiceIdExist(serviceId service.ServiceID) bool {
	_, ok := environment.serviceIDs[serviceId]
	return ok
}

func (environment *ValidatorEnvironment) AddArtifactId(artifactId enclave_data_directory.FilesArtifactID) {
	environment.artifactIDs[artifactId] = true
}

func (environment *ValidatorEnvironment) RemoveArtifactId(artifactId enclave_data_directory.FilesArtifactID) {
	delete(environment.artifactIDs, artifactId)
}

func (environment *ValidatorEnvironment) DoesArtifactIdExist(artifactId enclave_data_directory.FilesArtifactID) bool {
	_, ok := environment.artifactIDs[artifactId]
	return ok
}
