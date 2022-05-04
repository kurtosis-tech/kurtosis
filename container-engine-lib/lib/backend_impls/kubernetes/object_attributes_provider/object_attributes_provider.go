package object_attributes_provider

type KubernetesObjectAttributesProvider interface {
	ForEngine(
		id string,
	) (KubernetesEngineObjectAttributesProvider, error)
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
