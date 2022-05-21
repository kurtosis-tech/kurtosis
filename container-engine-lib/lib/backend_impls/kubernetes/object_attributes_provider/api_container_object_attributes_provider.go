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
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// All API container objects will be named this, which is fine because the objects are namespaced and there
	// should only be a single API container per enclave
	apiContainerObjectNameStr                = "kurtosis-api"
)
var apiContainerObjectName = kubernetes_object_name.MustCreateNewKubernetesObjectName(apiContainerObjectNameStr)

type KubernetesApiContainerObjectAttributesProvider interface {
	ForApiContainerPod() (KubernetesObjectAttributes, error)
	ForApiContainerService(
		privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
		privateGrpcProxyPortId string,
		privateGrpcProxyPortSpec *port_spec.PortSpec) (KubernetesObjectAttributes, error)
	ForApiContainerServiceAccount() (KubernetesObjectAttributes, error)
	ForApiContainerRole() (KubernetesObjectAttributes, error)
	ForApiContainerRoleBindings() (KubernetesObjectAttributes, error)
}

// Private so it can't be instantiated
type kubernetesApiContainerObjectAttributesProviderImpl struct {
	enclaveId string
}

func GetKubernetesApiContainerObjectAttributesProvider(enclaveId enclave.EnclaveID) KubernetesApiContainerObjectAttributesProvider {
	return newKubernetesApiContainerObjectAttributesProviderImpl(enclaveId)
}

func newKubernetesApiContainerObjectAttributesProviderImpl(enclaveId enclave.EnclaveID) *kubernetesApiContainerObjectAttributesProviderImpl {
	return &kubernetesApiContainerObjectAttributesProviderImpl{
		enclaveId: string(enclaveId),
	}
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) ForApiContainerPod() (KubernetesObjectAttributes, error) {
	labels, err := provider.getLabelsForApiContainerObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for API container object in enclave with ID '%v'", provider.enclaveId)
	}

	// No custom annotations for API container pod
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(apiContainerObjectName, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", apiContainerObjectName.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) ForApiContainerService(
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	labels, err := provider.getLabelsForApiContainerObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for API container object in enclave with ID '%v'", provider.enclaveId)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		grpcPortId:      grpcPortSpec,
		grpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := kubernetes_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following API container server ports to a string for storing in the ports annotation: %+v", usedPorts)
	}

	// Store Kurtosis port_spec info in annotation
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		annotation_key_consts.PortSpecsKubernetesAnnotationKey: serializedPortsSpec,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(apiContainerObjectName, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", apiContainerObjectName.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) ForApiContainerServiceAccount() (KubernetesObjectAttributes, error) {
	labels, err := provider.getLabelsForApiContainerObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for API container object in enclave with ID '%v'", provider.enclaveId)
	}

	// No custom annotations for API container service account
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(apiContainerObjectName, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", apiContainerObjectName.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) ForApiContainerRole() (KubernetesObjectAttributes, error) {
	labels, err := provider.getLabelsForApiContainerObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for API container object in enclave with ID '%v'", provider.enclaveId)
	}

	// No custom annotations for API container role
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(apiContainerObjectName, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", apiContainerObjectName.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) ForApiContainerRoleBindings() (KubernetesObjectAttributes, error) {
	labels, err := provider.getLabelsForApiContainerObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for API container object in enclave with ID '%v'", provider.enclaveId)
	}

	// No custom annotations for API container role bindings
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(apiContainerObjectName, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name " +
			"'%s' and labels '%+v', and annotations '%+v'", apiContainerObjectName.GetString(), labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesApiContainerObjectAttributesProviderImpl) getLabelsForApiContainerObject() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create Kubernetes label value from enclaveId '%v'", provider.enclaveId)
	}
	return map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.APIContainerKurtosisResourceTypeKubernetesLabelValue,
		label_key_consts.EnclaveIDKubernetesLabelKey:            enclaveIdLabelValue,
	}, nil
}
