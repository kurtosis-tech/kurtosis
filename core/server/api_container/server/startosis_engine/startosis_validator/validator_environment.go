package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	requiredDockerImages map[string]bool
	serviceIDs           map[service.ServiceID]bool
	artifactUUIDs        map[enclave_data_directory.FilesArtifactUUID]bool
}

func NewValidatorEnvironment(serviceIDs map[service.ServiceID]bool) *ValidatorEnvironment {
	return &ValidatorEnvironment{
		requiredDockerImages: map[string]bool{},
		serviceIDs:           serviceIDs,
		artifactUUIDs:        map[enclave_data_directory.FilesArtifactUUID]bool{},
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

func (environment *ValidatorEnvironment) AddArtifactUuid(artifactUuid enclave_data_directory.FilesArtifactUUID) {
	environment.artifactUUIDs[artifactUuid] = true
}

func (environment *ValidatorEnvironment) RemoveArtifactUuid(artifactUuid enclave_data_directory.FilesArtifactUUID) {
	delete(environment.artifactUUIDs, artifactUuid)
}

func (environment *ValidatorEnvironment) DoesArtifactUuidExist(artifactUuid enclave_data_directory.FilesArtifactUUID) bool {
	_, ok := environment.artifactUUIDs[artifactUuid]
	return ok
}
