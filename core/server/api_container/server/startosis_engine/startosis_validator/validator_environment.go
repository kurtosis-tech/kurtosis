package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/sirupsen/logrus"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled bool
	requiredDockerImages         map[string]bool
	serviceNames                 map[service.ServiceName]bool
	artifactNames                map[string]bool
	serviceNameToPrivatePortIDs  map[service.ServiceName][]string
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, artifactNames map[string]bool) *ValidatorEnvironment {
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled: isNetworkPartitioningEnabled,
		requiredDockerImages:         map[string]bool{},
		serviceNames:                 map[service.ServiceName]bool{},
		artifactNames:                artifactNames,
		serviceNameToPrivatePortIDs:  map[service.ServiceName][]string{},
	}
}

func (environment *ValidatorEnvironment) AppendRequiredContainerImage(containerImage string) {
	environment.requiredDockerImages[containerImage] = true
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImages() uint32 {
	return uint32(len(environment.requiredDockerImages))
}

func (environment *ValidatorEnvironment) AddServiceName(serviceName service.ServiceName) {
	logrus.Debugf("Adding service '%s' to validation environment", serviceName)
	environment.serviceNames[serviceName] = true
}

func (environment *ValidatorEnvironment) RemoveServiceName(serviceName service.ServiceName) {
	logrus.Debugf("Removing service '%s' from validation environment", serviceName)
	delete(environment.serviceNames, serviceName)
}

func (environment *ValidatorEnvironment) DoesServiceNameExist(serviceName service.ServiceName) bool {
	_, ok := environment.serviceNames[serviceName]
	return ok
}

func (environment *ValidatorEnvironment) AddPrivatePortIDForService(portID string, serviceName service.ServiceName) {
	logrus.Debugf("Adding private port ID '%s' for service '%s' to validation environment", serviceName, portID)
	environment.serviceNameToPrivatePortIDs[serviceName] = append(environment.serviceNameToPrivatePortIDs[serviceName], portID)
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
	logrus.Debugf("Removing all private ports IDs for service '%s' from validation environment", serviceName)
	delete(environment.serviceNameToPrivatePortIDs, serviceName)
}

func (environment *ValidatorEnvironment) AddArtifactName(artifactName string) {
	logrus.Debugf("Adding artifact name '%s' to validation environment", artifactName)
	environment.artifactNames[artifactName] = true
}

func (environment *ValidatorEnvironment) RemoveArtifactName(artifactName string) {
	logrus.Debugf("Removing artifact name '%s' from validation environment", artifactName)
	delete(environment.artifactNames, artifactName)
}

func (environment *ValidatorEnvironment) DoesArtifactNameExist(artifactName string) bool {
	_, ok := environment.artifactNames[artifactName]
	return ok
}

func (environment *ValidatorEnvironment) IsNetworkPartitioningEnabled() bool {
	return environment.isNetworkPartitioningEnabled
}
