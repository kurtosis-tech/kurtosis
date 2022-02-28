package kurtosis_docker_driver

// A backend that provides Kurtosis-flavored verbs and does the necessary work with the Docker backend to
//  accomplish what the user desires
type KurtosisDockerDriver struct {

}

func (manager *KurtosisDockerDriver) CreateEngineContainer() {}
func (manager *KurtosisDockerDriver) GetEngineContainers()   {}

// func (manager *KurtosisDockerDriver) CreateEnclaveNetwork() {}
// func (manager *KurtosisDockerDriver) GetEnclaveNetworks() {}
// func (manager *KurtosisDockerDriver) CreateEnclaveDataVolume() {}
// func (manager *KurtosisDockerDriver) GetEnclaveDataVolumes() {}
// func (manager *KurtosisDockerDriver) CreateAPIContainer() {}
// func (manager *KurtosisDockerDriver) GetAPIContainers() {}
// // Used by both Core (in CoreBackend.GetServices() and CliBackend.GetServices(), because you want to be able
// // to display the services inside an enclave even if the API container is stopped
// func (manager *KurtosisDockerDriver) CreateUserServiceContainer(usedPorts map[string]PortSpec) {}
// func (manager *KurtosisDockerDriver) GetUserServiceContainers() {}
// func (manager *KurtosisDockerDriver) CreateNetworkingSidecarContainer() {}
// func (manager *KurtosisDockerDriver) GetNetworkingSidecarContainers() {}
// func (manager *KurtosisDockerDriver) CreateFilesArtifactExpansionVolume() {}
// func (manager *KurtosisDockerDriver) GetFilesArtifactExpansionVolumes() {}
// func (manager *KurtosisDockerDriver) CreateFilesArtifactExpansionContainer() {}
// func (manager *KurtosisDockerDriver) GetFilesArtifactExpansionContainers() {}

func (manager *KurtosisDockerDriver) StopContainers()    {}
func (manager *KurtosisDockerDriver) DestroyContainers() {}
func (manager *KurtosisDockerDriver) DestroyVolumes()    {}
func (manager *KurtosisDockerDriver) DestroyNetworks()   {}