package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	namespacePrefix   = "kurtosis-enclave"
	userServicePrefix = "user-service"
)

type KubernetesEnclaveObjectAttributesProvider interface {
	ForEnclaveNamespace(isPartitioningEnabled bool, creationTime time.Time, enclaveName string) (KubernetesObjectAttributes, error)
	ForApiContainer() KubernetesApiContainerObjectAttributesProvider
	ForUserServiceService(
		uuid service.ServiceUUID,
		id service.ServiceName,
	) (KubernetesObjectAttributes, error)
	ForUserServicePod(
		uuid service.ServiceUUID,
		id service.ServiceName,
		privatePorts map[string]*port_spec.PortSpec,
	) (KubernetesObjectAttributes, error)
}

// Private so it can't be instantiated
type kubernetesEnclaveObjectAttributesProviderImpl struct {
	enclaveId string
}

func newKubernetesEnclaveObjectAttributesProviderImpl(

	enclaveId enclave.EnclaveUUID,
) *kubernetesEnclaveObjectAttributesProviderImpl {
	return &kubernetesEnclaveObjectAttributesProviderImpl{
		enclaveId: string(enclaveId),
	}
}

func GetKubernetesEnclaveObjectAttributesProvider(enclaveId enclave.EnclaveUUID) KubernetesEnclaveObjectAttributesProvider {
	return newKubernetesEnclaveObjectAttributesProviderImpl(enclaveId)
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveNamespace(isPartitioningEnabled bool, creationTime time.Time, enclaveName string) (KubernetesObjectAttributes, error) {
	namespaceUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to generate UUID string for namespace name for enclave '%v'", provider.enclaveId)
	}
	name, err := getCompositeKubernetesObjectName([]string{
		namespacePrefix,
		namespaceUuid,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", provider.enclaveId)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(
		provider.enclaveId,
		provider.enclaveId,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for enclave namespace using ID '%v'", provider.enclaveId)
	}

	isPartitioningEnabledLabelValue := label_value_consts.NetworkPartitioningDisabledKubernetesLabelValue
	if isPartitioningEnabled {
		isPartitioningEnabledLabelValue = label_value_consts.NetworkPartitioningEnabledKubernetesLabelValue
	}

	labels[label_key_consts.IsNetworkPartitioningEnabledKubernetesLabelKey] = isPartitioningEnabledLabelValue

	creationTimeStr := creationTime.Format(time.RFC3339)

	creationTimeAnnotationValue, err := kubernetes_annotation_value.CreateNewKubernetesAnnotationValue(creationTimeStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes annotation value from string '%v'", creationTimeStr)
	}

	enclaveNameAnnotationValue, err := kubernetes_annotation_value.CreateNewKubernetesAnnotationValue(enclaveName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes annotation value from string '%v'", enclaveName)
	}

	// Store enclave's creation time info in annotation
	customAnnotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key_consts.EnclaveCreationTimeAnnotationKey: creationTimeAnnotationValue,
		kubernetes_annotation_key_consts.EnclaveNameAnnotationKey:         enclaveNameAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, customAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the Kubernetes object attributes impl with the name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForApiContainer() KubernetesApiContainerObjectAttributesProvider {
	enclaveId := enclave.EnclaveUUID(provider.enclaveId)
	return GetKubernetesApiContainerObjectAttributesProvider(enclaveId)
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForNetworkingSidecarContainer(
	serviceUUIDSidecarAttachedTo service.ServiceUUID,
) (KubernetesObjectAttributes, error) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServiceService(
	serviceUUID service.ServiceUUID,
	serviceName service.ServiceName,
) (
	KubernetesObjectAttributes,
	error,
) {
	name, err := getCompositeKubernetesObjectName([]string{
		userServicePrefix,
		string(serviceUUID),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service service.")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(serviceName), string(serviceUUID))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service service with ID '%s' and UUID '%s'.",
			string(serviceName),
			string(serviceUUID),
		)
	}
	labels[label_key_consts.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	//No userServiceService annotations.
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service service object attributes.")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServicePod(
	uuid service.ServiceUUID,
	id service.ServiceName,
	privatePorts map[string]*port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{
		userServicePrefix,
		string(uuid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service pod")
	}

	serializedPortSpecsAnnotationValue, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following user service port specs to a string for storing in the ports label: %+v", privatePorts)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(id), string(uuid))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service pod with ID '%s' and UUID '%s'",
			id,
			uuid,
		)
	}
	labels[label_key_consts.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey: serializedPortSpecsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service pod object attributes")
	}

	return objectAttributes, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create Kubernetes label value from enclaveId '%v'", provider.enclaveId)
	}
	return map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue,
		label_key_consts.EnclaveUUIDKubernetesLabelKey:          enclaveIdLabelValue,
	}, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithUUID(uuid string) (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for enclave object with uuid '%v'", uuid)
	}
	uuidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(uuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes label value from UUID string '%v'", uuid)
	}
	labels[label_key_consts.GUIDKubernetesLabelKey] = uuidLabelValue
	return labels, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithIDAndGUID(id, uuid string) (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObjectWithUUID(uuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave object labels with GUID '%v'", uuid)
	}
	idLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes label value from ID string '%v'", id)
	}
	labels[label_key_consts.IDKubernetesLabelKey] = idLabelValue
	return labels, nil
}

func getLabelKeyValuesAsStrings(labels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range labels {
		result[key.GetString()] = value.GetString()
	}
	return result
}
