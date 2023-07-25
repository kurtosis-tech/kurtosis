package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled bool
	requiredDockerImages         map[string]bool
	serviceNames                 map[service.ServiceName]ComponentExistence
	artifactNames                map[string]ComponentExistence
	serviceNameToPrivatePortIDs  map[service.ServiceName][]string
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, serviceNames map[service.ServiceName]bool, artifactNames map[string]bool, serviceNameToPrivatePortIds map[service.ServiceName][]string) *ValidatorEnvironment {
	serviceNamesWithExistence := map[service.ServiceName]ComponentExistence{}
	for serviceName := range serviceNames {
		serviceNamesWithExistence[serviceName] = ComponentExistedBeforePackageRun
	}
	artifactNamesWithExistence := map[string]ComponentExistence{}
	for artifactName := range artifactNames {
		artifactNamesWithExistence[artifactName] = ComponentExistedBeforePackageRun
	}
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled: isNetworkPartitioningEnabled,
		requiredDockerImages:         map[string]bool{},
		serviceNames:                 serviceNamesWithExistence,
		artifactNames:                artifactNamesWithExistence,
		serviceNameToPrivatePortIDs:  serviceNameToPrivatePortIds,
	}
}

func (environment *ValidatorEnvironment) AppendRequiredContainerImage(containerImage string) {
	environment.requiredDockerImages[containerImage] = true
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImages() uint32 {
	return uint32(len(environment.requiredDockerImages))
}

func (environment *ValidatorEnvironment) AddServiceName(serviceName service.ServiceName) {
	environment.serviceNames[serviceName] = ComponentCreatedOrUpdatedDuringPackageRun
}

func (environment *ValidatorEnvironment) RemoveServiceName(serviceName service.ServiceName) {
	delete(environment.serviceNames, serviceName)
}

func (environment *ValidatorEnvironment) DoesServiceNameExist(serviceName service.ServiceName) ComponentExistence {
	serviceExistence, found := environment.serviceNames[serviceName]
	if !found {
		return ComponentNotFound
	}
	return serviceExistence
}

func (environment *ValidatorEnvironment) AddPrivatePortIDForService(portIDs []string, serviceName service.ServiceName) {
	environment.serviceNameToPrivatePortIDs[serviceName] = portIDs
}

func (environment *ValidatorEnvironment) DoesPrivatePortIDExistForService(portID string, serviceName service.ServiceName) bool {
	existingPortIDs, found := environment.serviceNameToPrivatePortIDs[serviceName]
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
	delete(environment.serviceNameToPrivatePortIDs, serviceName)
}

func (environment *ValidatorEnvironment) AddArtifactName(artifactName string) {
	environment.artifactNames[artifactName] = ComponentCreatedOrUpdatedDuringPackageRun
}

func (environment *ValidatorEnvironment) RemoveArtifactName(artifactName string) {
	delete(environment.artifactNames, artifactName)
}

func (environment *ValidatorEnvironment) DoesArtifactNameExist(artifactName string) ComponentExistence {
	artifactExistence, found := environment.artifactNames[artifactName]
	if !found {
		return ComponentNotFound
	}
	return artifactExistence
}

func (environment *ValidatorEnvironment) IsNetworkPartitioningEnabled() bool {
	return environment.isNetworkPartitioningEnabled
}
