package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	isNetworkPartitioningEnabled  bool
	requiredDockerImages          map[string]bool
	serviceNames                  map[service.ServiceName]ServiceExistence
	artifactNames                 map[string]bool
	serviceNameToPrivatePortIDs   map[service.ServiceName][]string
	availableCpuInMilliCores      compute_resources.CpuMilliCores
	availableMemoryInMegaBytes    compute_resources.MemoryInMegaBytes
	isResourceInformationComplete bool
	minCPUByServiceName           map[service.ServiceName]compute_resources.CpuMilliCores
	minMemoryByServiceName        map[service.ServiceName]compute_resources.MemoryInMegaBytes
}

func NewValidatorEnvironment(isNetworkPartitioningEnabled bool, serviceNames map[service.ServiceName]bool, artifactNames map[string]bool, serviceNameToPrivatePortIds map[service.ServiceName][]string, availableCpuInMilliCores compute_resources.CpuMilliCores, availableMemoryInMegaBytes compute_resources.MemoryInMegaBytes, isResourceInformationComplete bool) *ValidatorEnvironment {
	serviceNamesWithServiceExistence := map[service.ServiceName]ServiceExistence{}
	for serviceName := range serviceNames {
		serviceNamesWithServiceExistence[serviceName] = ServiceExistedBeforePackageRun
	}
	return &ValidatorEnvironment{
		isNetworkPartitioningEnabled:  isNetworkPartitioningEnabled,
		requiredDockerImages:          map[string]bool{},
		serviceNames:                  serviceNamesWithServiceExistence,
		artifactNames:                 artifactNames,
		serviceNameToPrivatePortIDs:   serviceNameToPrivatePortIds,
		availableCpuInMilliCores:      availableCpuInMilliCores,
		availableMemoryInMegaBytes:    availableMemoryInMegaBytes,
		isResourceInformationComplete: isResourceInformationComplete,
		minMemoryByServiceName:        map[service.ServiceName]compute_resources.MemoryInMegaBytes{},
		minCPUByServiceName:           map[service.ServiceName]compute_resources.CpuMilliCores{},
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

func (environment *ValidatorEnvironment) FreeMemory(serviceName service.ServiceName) {
	memoryConsumedByService, found := environment.minMemoryByServiceName[serviceName]
	if !found {
		logrus.Warnf("tried to run 'FreeMemory' for service '%v' that didn't exist in validator", serviceName)
		return
	}
	environment.availableMemoryInMegaBytes += memoryConsumedByService
}

func (environment *ValidatorEnvironment) ConsumeMemory(memoryConsumed uint64, serviceName service.ServiceName) {
	environment.availableMemoryInMegaBytes -= compute_resources.MemoryInMegaBytes(memoryConsumed)
	environment.minMemoryByServiceName[serviceName] = compute_resources.MemoryInMegaBytes(memoryConsumed)
}

func (environment *ValidatorEnvironment) FreeCPU(serviceName service.ServiceName) {
	cpuConsumedByService, found := environment.minCPUByServiceName[serviceName]
	if !found {
		logrus.Warnf("tried to run 'FreeCPU' for service '%v' that didn't exist in validator", serviceName)
		return
	}
	environment.availableCpuInMilliCores += cpuConsumedByService
}

func (environment *ValidatorEnvironment) ConsumeCPU(cpuConsumed uint64, serviceName service.ServiceName) {
	environment.availableCpuInMilliCores -= compute_resources.CpuMilliCores(cpuConsumed)
	environment.minCPUByServiceName[serviceName] = compute_resources.CpuMilliCores(cpuConsumed)
}

func (environment *ValidatorEnvironment) HasEnoughCPU(cpuToConsume uint64, serviceNameForLogging service.ServiceName) *startosis_errors.ValidationError {
	if !environment.isResourceInformationComplete {
		return nil
	}
	if environment.availableCpuInMilliCores >= compute_resources.CpuMilliCores(cpuToConsume) {
		return nil
	}
	return startosis_errors.NewValidationError("service '%v' requires '%v' millicores of cpu but based on our calculation we will only have '%v' millicores available at the time we start the service", serviceNameForLogging, cpuToConsume, environment.availableCpuInMilliCores)
}

func (environment *ValidatorEnvironment) HasEnoughMemory(memoryToConsume uint64, serviceNameForLogging service.ServiceName) *startosis_errors.ValidationError {
	if !environment.isResourceInformationComplete {
		return nil
	}
	if environment.availableMemoryInMegaBytes >= compute_resources.MemoryInMegaBytes(memoryToConsume) {
		return nil
	}
	return startosis_errors.NewValidationError("service '%v' requires '%v' megabytes of memory but based on our calculation we will only have '%v' megabytes available at the time we start the service", serviceNameForLogging, memoryToConsume, environment.availableMemoryInMegaBytes)
}
