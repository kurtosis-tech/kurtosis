package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	engineNamePrefix        = "kurtosis-engine"
	enginePodNameSuffix     = "pod"
	engineServiceNameSuffix = "service"
)

type KubernetesEngineObjectAttributesProvider interface {
	ForEnginePod() (KubernetesObjectAttributes, error)

	ForEngineService(privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
		privateGrpcProxyPortId string,
		privateGrpcProxyPortSpec *port_spec.PortSpec) (KubernetesObjectAttributes, error)
}

// Private so it can't be instantiated
type kubernetesEngineObjectAttributesProviderImpl struct {
	engineId string
}

func GetKubernetesEngineObjectAttributesProvider(enclaveId string) KubernetesEngineObjectAttributesProvider {
	return newKubernetesEngineObjectAttributesProviderImpl(enclaveId)
}

func newKubernetesEngineObjectAttributesProviderImpl(
	engineId string,
) *kubernetesEngineObjectAttributesProviderImpl {
	return &kubernetesEngineObjectAttributesProviderImpl{
		engineId: engineId,
	}
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEnginePod() (KubernetesObjectAttributes, error) {
	nameStr := provider.getEngineObjectNameString(enginePodNameSuffix)
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name object from string '%v'", nameStr)
	}

	idLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.engineId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine ID Kubernetes label from string '%v'", provider.engineId)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.ResourceTypeLabelKey: label_value_consts.EngineResourceTypeLabelValue,
		label_key_consts.IDLabelKey:           idLabelValue,
	}

	// No custom annotations for engine pod
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name '%s' and labels '%+v', and annotations '%+v'", name, labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineService(grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	nameStr := provider.getEngineObjectNameString(engineServiceNameSuffix)
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name for our engine service")
	}

	idLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.engineId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine ID Kubernetes label from string '%v'", provider.engineId)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.ResourceTypeLabelKey: label_value_consts.EngineResourceTypeLabelValue,
		label_key_consts.IDLabelKey:           idLabelValue,
	}

	usedPorts := map[string]*port_spec.PortSpec{
		grpcPortId:      grpcPortSpec,
		grpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports annotation: %+v", usedPorts)
	}

	// Store Kurtosis port_spec info in annotation
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		annotation_key_consts.PortSpecsAnnotationKey: serializedPortsSpec,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name '%s' and labels '%+v', and annotations '%+v'", name, labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) GetEngineSelectorLabels() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	idLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.engineId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine ID Kubernetes label from string '%v'", provider.engineId)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.ResourceTypeLabelKey: label_value_consts.EngineResourceTypeLabelValue,
		label_key_consts.IDLabelKey:           idLabelValue,
	}

	return labels, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) getEngineObjectNameString(suffix string) string {
	nameStr := strings.Join(
		[]string{
			engineNamePrefix,
			provider.engineId,
			suffix,
		},
		objectNameElementSeparator,
	)
	return nameStr

}
