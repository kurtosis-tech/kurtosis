package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type KubernetesObjectAttributesProvider interface {
	ForEngine(id string) KubernetesEngineObjectAttributesProvider
	ForEnclave(enclaveId enclave.EnclaveID) KubernetesEnclaveObjectAttributesProvider
}

func GetKubernetesObjectAttributesProvider() KubernetesObjectAttributesProvider {
	return newKubernetesObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type kubernetesObjectAttributesProviderImpl struct{}

func newKubernetesObjectAttributesProviderImpl() *kubernetesObjectAttributesProviderImpl {
	return &kubernetesObjectAttributesProviderImpl{}
}

func (provider *kubernetesObjectAttributesProviderImpl) ForEngine(engineId string) KubernetesEngineObjectAttributesProvider {
	return GetKubernetesEngineObjectAttributesProvider(engineId)
}

func (provider *kubernetesObjectAttributesProviderImpl) ForEnclave(enclaveId enclave.EnclaveID) KubernetesEnclaveObjectAttributesProvider {
	return GetKubernetesEnclaveObjectAttributesProvider(enclaveId)
}

