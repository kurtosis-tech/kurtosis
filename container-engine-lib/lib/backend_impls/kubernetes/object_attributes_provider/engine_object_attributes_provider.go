package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	engineNamePrefix                = "kurtosis-engine"
)

type KubernetesEngineObjectAttributesProvider interface {
	ForEnginePod() (KubernetesObjectAttributes, error)

	ForEngineService(privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
		privateGrpcProxyPortId string,
		privateGrpcProxyPortSpec *port_spec.PortSpec) (KubernetesObjectAttributes, error)

	ForEngineNamespace() (KubernetesObjectAttributes, error)

	ForEngineServiceAccount() (KubernetesObjectAttributes, error)

	ForEngineClusterRole() (KubernetesObjectAttributes, error)

	ForEngineClusterRoleBindings() (KubernetesObjectAttributes, error)
}

// Private so it can't be instantiated
type kubernetesEngineObjectAttributesProviderImpl struct {
	engineGuid engine.EngineGUID
}

func GetKubernetesEngineObjectAttributesProvider(engineGuid engine.EngineGUID) KubernetesEngineObjectAttributesProvider {
	return newKubernetesEngineObjectAttributesProviderImpl(engineGuid)
}

func newKubernetesEngineObjectAttributesProviderImpl(
	engineGuid engine.EngineGUID,
) *kubernetesEngineObjectAttributesProviderImpl {
	return &kubernetesEngineObjectAttributesProviderImpl{
		engineGuid: engineGuid,
	}
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEnginePod() (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a pod name for the engine pod")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine pod
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineService(grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name for the engine service")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	usedPorts := map[string]*port_spec.PortSpec{
		grpcPortId:      grpcPortSpec,
		grpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := kubernetes_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports annotation: %+v", usedPorts)
	}

	// Store Kurtosis port_spec info in annotation
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		annotation_key_consts.PortSpecsAnnotationKey: serializedPortsSpec,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineNamespace() (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name for the engine namespace")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine namespace
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineServiceAccount() (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name for the engine service account")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine service account
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineClusterRole() (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name for the engine cluster role")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine cluster role
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) ForEngineClusterRoleBindings() (KubernetesObjectAttributes, error) {
	name, err := provider.getEngineObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes object name for the engine cluster role binding")
	}

	labels, err := provider.getEngineObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine cluster role bindings
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================
func (provider *kubernetesEngineObjectAttributesProviderImpl) getEngineObjectLabels() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	guidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(provider.engineGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing engine GUID '%v' into a Kubernetes label value", provider.engineGuid)
	}

	// ID and GUID are the same here
	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.EngineKurtosisResourceTypeKubernetesLabelValue,
		label_key_consts.IDKubernetesLabelKey:                   guidLabelValue,
		label_key_consts.GUIDKubernetesLabelKey:                   guidLabelValue,
	}
	return labels, nil
}

func (provider *kubernetesEngineObjectAttributesProviderImpl) getEngineObjectName() (*kubernetes_object_name.KubernetesObjectName, error) {
	result, err := getCompositeKubernetesObjectName([]string{
		engineNamePrefix,
		string(provider.engineGuid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a Kubernetes object name for engine GUID '%v'", provider.engineGuid)
	}
	return result, nil
}
