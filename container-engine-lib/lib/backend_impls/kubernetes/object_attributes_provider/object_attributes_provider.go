package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

type KubernetesObjectAttributesProvider interface {
	ForEngine(guid engine.EngineGUID) KubernetesEngineObjectAttributesProvider
	ForEnclave(enclaveId enclave.EnclaveUUID) KubernetesEnclaveObjectAttributesProvider
	ForLogsCollector(guid logs_collector.LogsCollectorGuid) KubernetesLogsCollectorObjectAttributesProvider
}

func GetKubernetesObjectAttributesProvider() KubernetesObjectAttributesProvider {
	return newKubernetesObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type kubernetesObjectAttributesProviderImpl struct{}

func newKubernetesObjectAttributesProviderImpl() *kubernetesObjectAttributesProviderImpl {
	return &kubernetesObjectAttributesProviderImpl{}
}

func (provider *kubernetesObjectAttributesProviderImpl) ForEngine(engineGuid engine.EngineGUID) KubernetesEngineObjectAttributesProvider {
	return GetKubernetesEngineObjectAttributesProvider(engineGuid)
}

func (provider *kubernetesObjectAttributesProviderImpl) ForEnclave(enclaveId enclave.EnclaveUUID) KubernetesEnclaveObjectAttributesProvider {
	return GetKubernetesEnclaveObjectAttributesProvider(enclaveId)
}

func (provider *kubernetesObjectAttributesProviderImpl) ForLogsCollector(logsCollectorGuid logs_collector.LogsCollectorGuid) KubernetesLogsCollectorObjectAttributesProvider {
	return GetKubernetesLogsCollectorObjectAttributesProvider(logsCollectorGuid)
}

// Gets the name for an enclave object, making sure to put the enclave ID first and join using the standardized separator
func getCompositeKubernetesObjectName(elems []string) (*kubernetes_object_name.KubernetesObjectName, error) {
	nameStr := strings.Join(
		elems,
		objectNameElementSeparator,
	)
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes object name from string '%v'", nameStr)
	}
	return name, nil
}
