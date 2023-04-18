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
	serviceNameToPrivatePortIds  map[service.ServiceName][]string
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

func (environment *ValidatorEnvironment) AddServiceName(serviceName service.ServiceName) {
	environment.serviceNames[serviceName] = true
}

func (environment *ValidatorEnvironment) RemoveServiceName(serviceName service.ServiceName) {
	delete(environment.serviceNames, serviceName)
}

func (environment *ValidatorEnvironment) DoesServiceNameExist(serviceName service.ServiceName) bool {
	_, ok := environment.serviceNames[serviceName]
	return ok
}

func (environment *ValidatorEnvironment) AddPrivatePortIDForService(portID string, serviceName service.ServiceName) {
	portIds, found := environment.serviceNameToPrivatePortIds[serviceName]
	if found {
		environment.serviceNameToPrivatePortIds[serviceName] = append(portIds, portID)
		return
	}
	environment.serviceNameToPrivatePortIds[serviceName] = []string{portID}
}

func (environment *ValidatorEnvironment) DoesPrivatePortIDExistForService(portID string, serviceName service.ServiceName) bool {
	existingPortIDs, found := environment.serviceNameToPrivatePortIds[serviceName]
	if !found {
		return false
	}
	for _, existingPortID := range existingPortIDs {
		if existingPortID == portID {
			return true
		}
	}
	return false
}

func (environment *ValidatorEnvironment) RemoveServiceFromPrivatePortIDMapping(serviceName service.ServiceName) {
	delete(environment.serviceNameToPrivatePortIds, serviceName)
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
