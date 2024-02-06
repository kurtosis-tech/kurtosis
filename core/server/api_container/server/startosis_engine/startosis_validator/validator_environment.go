package startosis_validator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/compute_resources"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
)

// ValidatorEnvironment fields are not exported so that only validators can access its fields
type ValidatorEnvironment struct {
	imagesToPull                  map[string]*image_registry_spec.ImageRegistrySpec // "set" of images that need to be downloaded
	imagesToBuild                 map[string]*image_build_spec.ImageBuildSpec
	serviceNames                  map[service.ServiceName]ComponentExistence
	artifactNames                 map[string]ComponentExistence
	persistentKeys                map[service_directory.DirectoryPersistentKey]ComponentExistence
	serviceNameToPrivatePortIDs   map[service.ServiceName][]string
	availableCpuInMilliCores      compute_resources.CpuMilliCores
	availableMemoryInMegaBytes    compute_resources.MemoryInMegaBytes
	isResourceInformationComplete bool
	minCPUByServiceName           map[service.ServiceName]compute_resources.CpuMilliCores
	minMemoryByServiceName        map[service.ServiceName]compute_resources.MemoryInMegaBytes
	imageDownloadMode             image_download_mode.ImageDownloadMode
}

func NewValidatorEnvironment(serviceNames map[service.ServiceName]bool, artifactNames map[string]bool, serviceNameToPrivatePortIds map[service.ServiceName][]string, availableCpuInMilliCores compute_resources.CpuMilliCores, availableMemoryInMegaBytes compute_resources.MemoryInMegaBytes, isResourceInformationComplete bool, imageDownloadMode image_download_mode.ImageDownloadMode) *ValidatorEnvironment {
	serviceNamesWithComponentExistence := map[service.ServiceName]ComponentExistence{}
	for serviceName := range serviceNames {
		serviceNamesWithComponentExistence[serviceName] = ComponentExistedBeforePackageRun
	}
	artifactNamesWithComponentExistence := map[string]ComponentExistence{}
	for artifactName := range artifactNames {
		artifactNamesWithComponentExistence[artifactName] = ComponentExistedBeforePackageRun
	}
	return &ValidatorEnvironment{
		imagesToPull:                  map[string]*image_registry_spec.ImageRegistrySpec{},
		imagesToBuild:                 map[string]*image_build_spec.ImageBuildSpec{},
		serviceNames:                  serviceNamesWithComponentExistence,
		artifactNames:                 artifactNamesWithComponentExistence,
		serviceNameToPrivatePortIDs:   serviceNameToPrivatePortIds,
		availableCpuInMilliCores:      availableCpuInMilliCores,
		availableMemoryInMegaBytes:    availableMemoryInMegaBytes,
		isResourceInformationComplete: isResourceInformationComplete,
		// TODO account for idempotent runs on this and make it pre-load the cache whenever we create a NewValidatorEnvironment
		persistentKeys:         map[service_directory.DirectoryPersistentKey]ComponentExistence{},
		minMemoryByServiceName: map[service.ServiceName]compute_resources.MemoryInMegaBytes{},
		minCPUByServiceName:    map[service.ServiceName]compute_resources.CpuMilliCores{},
		imageDownloadMode:      imageDownloadMode,
	}
}

func (environment *ValidatorEnvironment) AppendRequiredImagePull(containerImage string) {
	environment.imagesToPull[containerImage] = nil
}

func (environment *ValidatorEnvironment) AppendRequiredImageBuild(containerImage string, imageBuildSpec *image_build_spec.ImageBuildSpec) {
	environment.imagesToBuild[containerImage] = imageBuildSpec
}

func (environmemt *ValidatorEnvironment) AppendImageToPullWithAuth(containerImage string, registrySpec *image_registry_spec.ImageRegistrySpec) {
	environmemt.imagesToPull[containerImage] = registrySpec
}

func (environment *ValidatorEnvironment) GetNumberOfContainerImagesToProcess() uint32 {
	return uint32(len(environment.imagesToPull) + len(environment.imagesToBuild))
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
	filesArtifactExistence, found := environment.artifactNames[artifactName]
	if !found {
		return ComponentNotFound
	}
	return filesArtifactExistence
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

func (environment *ValidatorEnvironment) AddPersistentKey(persistentKey service_directory.DirectoryPersistentKey) {
	environment.persistentKeys[persistentKey] = ComponentCreatedOrUpdatedDuringPackageRun
}
