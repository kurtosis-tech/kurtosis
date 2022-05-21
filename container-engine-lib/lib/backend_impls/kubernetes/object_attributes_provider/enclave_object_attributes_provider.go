package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	artifactExpansionObjectTimestampFormat = "2006-01-02T15.04.05.000"
	userServicePrefix = "user-service"
	modulePrefix      = "module"
)

type KubernetesEnclaveObjectAttributesProvider interface {
	ForEnclaveNamespace(isPartitioningEnabled bool) (KubernetesObjectAttributes, error)
	ForEnclaveDataPersistentVolumeClaim() (KubernetesObjectAttributes, error)
	ForApiContainer() KubernetesApiContainerObjectAttributesProvider
	ForUserServiceService(
		guid service.ServiceGUID,
		id service.ServiceID,
	) (KubernetesObjectAttributes, error)
	ForUserServicePod(
		guid service.ServiceGUID,
		id service.ServiceID,
		privatePorts map[string]*port_spec.PortSpec,
	) (KubernetesObjectAttributes, error)
	ForModulePod(
		guid module.ModuleGUID,
		id module.ModuleID,
		privatePorts map[string]*port_spec.PortSpec,
	) (KubernetesObjectAttributes, error)
	ForModuleService(
		guid module.ModuleGUID,
		id module.ModuleID,
		privatePorts map[string]*port_spec.PortSpec,
	) (KubernetesObjectAttributes, error)
}

// Private so it can't be instantiated
type kubernetesEnclaveObjectAttributesProviderImpl struct {
	enclaveId string
}

func newKubernetesEnclaveObjectAttributesProviderImpl(

	enclaveId enclave.EnclaveID,
) *kubernetesEnclaveObjectAttributesProviderImpl {
	return &kubernetesEnclaveObjectAttributesProviderImpl{
		enclaveId: string(enclaveId),
	}
}

func GetKubernetesEnclaveObjectAttributesProvider(enclaveId enclave.EnclaveID) KubernetesEnclaveObjectAttributesProvider {
	return newKubernetesEnclaveObjectAttributesProviderImpl(enclaveId)
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveNamespace(isPartitioningEnabled bool) (KubernetesObjectAttributes, error) {
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(provider.enclaveId)
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

	// No custom annotations for enclave namespace
	customAnnotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

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

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveDataPersistentVolumeClaim() (KubernetesObjectAttributes, error) {
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", provider.enclaveId)
	}

	labels, err := provider.getLabelsForEnclaveObject()
	labels[label_key_consts.KurtosisVolumeTypeKubernetesLabelKey] = label_value_consts.EnclaveDataVolumeTypeKubernetesLabelValue

	// No custom annotations for enclave data volume
	customAnnotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, customAnnotations)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the ObjectAttributesImpl with name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForApiContainer() KubernetesApiContainerObjectAttributesProvider{
	enclaveId := enclave.EnclaveID(provider.enclaveId)
	return GetKubernetesApiContainerObjectAttributesProvider(enclaveId)
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForNetworkingSidecarContainer(
	serviceGUIDSidecarAttachedTo service.ServiceGUID,
) (KubernetesObjectAttributes, error) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForFilesArtifactExpansionVolume(
	guid files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
)(
	KubernetesObjectAttributes,
	error,
){
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForFilesArtifactExpanderContainer(
	guid files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
)(
	KubernetesObjectAttributes,
	error,
) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServiceService (
	serviceGUID service.ServiceGUID,
	serviceID service.ServiceID,
) (
	KubernetesObjectAttributes,
	error,
) {
	name, err := getCompositeKubernetesObjectName([]string{
		userServicePrefix,
		string(serviceGUID),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service service.")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(serviceID), string(serviceGUID))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service service with ID '%s' and GUID '%s'.",
			string(serviceID),
			string(serviceGUID),
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
	guid service.ServiceGUID,
	id service.ServiceID,
	privatePorts map[string]*port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{
		userServicePrefix,
		string(guid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service pod")
	}

	serializedPortSpecsAnnotationValue, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following user service port specs to a string for storing in the ports label: %+v", privatePorts)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(id), string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service pod with ID '%s' and GUID '%s'",
			id,
			guid,
		)
	}
	labels[label_key_consts.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key.PortSpecsKubernetesAnnotationKey: serializedPortSpecsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service pod object attributes")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForModulePod(
	guid module.ModuleGUID,
	id module.ModuleID,
	privatePorts map[string]*port_spec.PortSpec,
) (
	KubernetesObjectAttributes,
	error,
) {
	name, err := getCompositeKubernetesObjectName([]string{
		modulePrefix,
		string(guid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for module pod.")
	}

	serializedPortSpecsAnnotationValue, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following module port specs to a string for storing in the ports label: %+v", privatePorts)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(id), string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for module pod with ID '%s' and GUID '%s'",
			id,
			guid,
		)
	}
	labels[label_key_consts.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.ModuleKurtosisResourceTypeKubernetesLabelValue

	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key.PortSpecsKubernetesAnnotationKey: serializedPortSpecsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name '%s' and labels '%+v', and annotations '%+v'", name, labels, annotations)
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForModuleService(
	guid module.ModuleGUID,
	id module.ModuleID,
	privatePorts map[string]*port_spec.PortSpec,
) (
	KubernetesObjectAttributes,
	error,
) {
	name, err := getCompositeKubernetesObjectName([]string{
		modulePrefix,
		string(guid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for module service.")
	}

	serializedPortSpecsAnnotationValue, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following module port specs to a string for storing in the ports label: %+v", privatePorts)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(id), string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for module service with ID '%s' and GUID '%s'",
			id,
			guid,
		)
	}
	labels[label_key_consts.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.ModuleKurtosisResourceTypeKubernetesLabelValue

	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key.PortSpecsKubernetesAnnotationKey: serializedPortSpecsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create module service object attributes.")
	}

	return objectAttributes, nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create Kubernetes label value from enclaveId '%v'", provider.enclaveId)
	}
	return map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue,
		label_key_consts.EnclaveIDKubernetesLabelKey:            enclaveIdLabelValue,
	}, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithGUID(guid string) (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for enclave object with guid '%v'", guid)
	}
	guidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kubernetes label value from GUID string '%v'", guid)
	}
	labels[label_key_consts.GUIDKubernetesLabelKey] = guidLabelValue
	return labels, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithIDAndGUID(id, guid string) (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave object labels with GUID '%v'", guid)
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