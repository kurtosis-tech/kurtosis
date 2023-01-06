package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled bool
	requiredDockerImages         map[string]bool
	serviceIDs                   map[service.ServiceID]bool
	artifactIDs                  map[enclave_data_directory.FilesArtifactID]bool
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, serviceIDs map[service.ServiceID]bool, artifactIDs map[enclave_data_directory.FilesArtifactID]bool) *ValidatorEnvironment {
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled: isNetworkPartitioningEnabled,
		requiredDockerImages:         map[string]bool{},
		serviceIDs:                   serviceIDs,
		artifactIDs:                  artifactIDs,
	}
}

func (environment *ValidatorEnvironment) AppendRequiredContainerImage(containerImage string) {
	environment.requiredDockerImages[containerImage] = true
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImages() uint32 {
	return uint32(len(environment.requiredDockerImages))
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

func (environment *ValidatorEnvironment) IsNetworkPartitioningEnabled() bool {
	return environment.isNetworkPartitioningEnabled
}
