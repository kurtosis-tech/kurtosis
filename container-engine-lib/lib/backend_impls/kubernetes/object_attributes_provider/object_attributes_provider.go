package object_attributes_provider

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"

type KubernetesObjectAttributesProvider interface {
	ForEngine(id string) (KubernetesEngineObjectAttributesProvider, error)
	ForApiContainer(enclaveId enclave.EnclaveID) (KubernetesApiContainerObjectAttributesProvider, error)
}

func GetKubernetesObjectAttributesProvider() KubernetesObjectAttributesProvider {
	return newKubernetesObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type kubernetesObjectAttributesProviderImpl struct{}

func newKubernetesObjectAttributesProviderImpl() *kubernetesObjectAttributesProviderImpl {
	return &kubernetesObjectAttributesProviderImpl{}
}

func (provider *kubernetesObjectAttributesProviderImpl) ForEngine(engineId string) (KubernetesEngineObjectAttributesProvider, error) {
	return GetKubernetesEngineObjectAttributesProvider(engineId), nil
}

func (provider *kubernetesObjectAttributesProviderImpl) ForApiContainer(enclaveId enclave.EnclaveID) (KubernetesApiContainerObjectAttributesProvider, error) {
	return GetKubernetesApiContainerObjectAttributesProvider(enclaveId), nil
}
