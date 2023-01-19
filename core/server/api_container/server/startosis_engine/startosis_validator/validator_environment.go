package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled bool
	requiredDockerImages         map[string]bool
	serviceNames                 map[service.ServiceName]bool
	artifactNames                map[string]bool
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, serviceNames map[service.ServiceName]bool, artifactNames map[string]bool) *ValidatorEnvironment {
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled: isNetworkPartitioningEnabled,
		requiredDockerImages:         map[string]bool{},
		serviceNames:                 serviceNames,
		artifactNames:                artifactNames,
	}
}

func (environment *ValidatorEnvironment) AppendRequiredContainerImage(containerImage string) {
	environment.requiredDockerImages[containerImage] = true
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImages() uint32 {
	return uint32(len(environment.requiredDockerImages))
}

func (environment *ValidatorEnvironment) AddServiceId(serviceId service.ServiceName) {
	environment.serviceNames[serviceId] = true
}

func (environment *ValidatorEnvironment) RemoveServiceId(serviceId service.ServiceName) {
	delete(environment.serviceNames, serviceId)
}

func (environment *ValidatorEnvironment) DoesServiceIdExist(serviceId service.ServiceName) bool {
	_, ok := environment.serviceNames[serviceId]
	return ok
}

func (environment *ValidatorEnvironment) AddArtifactName(artifactName string) {
	environment.artifactNames[artifactName] = true
}

func (environment *ValidatorEnvironment) RemoveArtifactName(artifactName string) {
	delete(environment.artifactNames, artifactName)
}

func (environment *ValidatorEnvironment) DoesArtifactNameExist(artifactName string) bool {
	_, ok := environment.artifactNames[artifactName]
	return ok
}

func (environment *ValidatorEnvironment) IsNetworkPartitioningEnabled() bool {
	return environment.isNetworkPartitioningEnabled
}
