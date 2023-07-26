package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled bool
	requiredDockerImages         map[string]bool
	serviceNames                 map[service.ServiceName]ServiceExistence
	artifactNames                map[string]bool
	serviceNameToPrivatePortIDs  map[service.ServiceName][]string
	availableCpuInMilliCores     uint64
	availableMemoryInMegaBytes   uint64
	skipCPUResourceCheck         bool
	skipMemoryResourceCheck      bool
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, serviceNames map[service.ServiceName]bool, artifactNames map[string]bool, serviceNameToPrivatePortIds map[service.ServiceName][]string, availableCpuInMilliCores uint64, availableMemoryInMegaBytes uint64, skipCPUResourceCheck bool, skipMemoryResourceCheck bool) *ValidatorEnvironment {
	serviceNamesWithServiceExistence := map[service.ServiceName]ServiceExistence{}
	for serviceName := range serviceNames {
		serviceNamesWithServiceExistence[serviceName] = ServiceExistedBeforePackageRun
	}
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled: isNetworkPartitioningEnabled,
		requiredDockerImages:         map[string]bool{},
		serviceNames:                 serviceNamesWithServiceExistence,
		artifactNames:                artifactNames,
		serviceNameToPrivatePortIDs:  serviceNameToPrivatePortIds,
		availableCpuInMilliCores:     availableCpuInMilliCores,
		availableMemoryInMegaBytes:   availableMemoryInMegaBytes,
		skipCPUResourceCheck:         skipCPUResourceCheck,
		skipMemoryResourceCheck:      skipMemoryResourceCheck,
	}
}

func (environment *ValidatorEnvironment) AppendRequiredContainerImage(containerImage string) {
	environment.requiredDockerImages[containerImage] = true
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImages() uint32 {
	return uint32(len(environment.requiredDockerImages))
}

func (environment *ValidatorEnvironment) AddServiceName(serviceName service.ServiceName) {
	environment.serviceNames[serviceName] = ServiceCreatedOrUpdatedDuringPackageRun
}

func (environment *ValidatorEnvironment) RemoveServiceName(serviceName service.ServiceName) {
	delete(environment.serviceNames, serviceName)
}

func (environment *ValidatorEnvironment) DoesServiceNameExist(serviceName service.ServiceName) ServiceExistence {
	serviceExistence, found := environment.serviceNames[serviceName]
	if !found {
		return ServiceNotFound
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

func (environment *ValidatorEnvironment) FreeMemory(memoryFreed uint64) {
	environment.availableMemoryInMegaBytes += memoryFreed
}

func (environment *ValidatorEnvironment) ConsumeMemory(memoryConsumed uint64) {
	environment.availableMemoryInMegaBytes -= memoryConsumed
}

func (environment *ValidatorEnvironment) FreeCPU(cpuFreed uint64) {
	environment.availableCpuInMilliCores += cpuFreed
}

func (environment *ValidatorEnvironment) ConsumeCPU(cpuConsumed uint64) {
	environment.availableCpuInMilliCores -= cpuConsumed
}

func (environment *ValidatorEnvironment) HasEnoughCPU(minCpuConsumed int, freeCPUInBackend int) {

}

func (environment *ValidatorEnvironment) HasEnoughMemory(minMemoryConsumed int, freeMemoryInBackend int) {

}
