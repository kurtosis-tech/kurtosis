package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strings"
	"time"
)

const (
	artifactExpansionObjectTimestampFormat = "2006-01-02T15.04.05.000"
	userServiceRegistrationSuffix = "kurtosis-user-service-registration"

)

type KubernetesEnclaveObjectAttributesProvider interface {
	ForEnclaveNamespace(isPartitioningEnabled bool) (KubernetesObjectAttributes, error)
	ForEnclaveDataVolume() (KubernetesObjectAttributes, error)
	ForApiContainer() (KubernetesApiContainerObjectAttributesProvider, error)
	ForUserServiceRegistration() (KubernetesObjectAttributes, error)
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

	labels, err := provider.getLabelsForEnclaveObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for enclave object '%v'", provider.enclaveId)
	}

	isPartitioningEnabledLabelValue := label_value_consts.NetworkPartitioningDisabledLabelValue
	if isPartitioningEnabled {
		isPartitioningEnabledLabelValue = label_value_consts.NetworkPartitioningEnabledLabelValue
	}

	labels[label_key_consts.IsNetworkPartitioningEnabledLabelKey] = isPartitioningEnabledLabelValue

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

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveDataVolume() (KubernetesObjectAttributes, error) {
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", provider.enclaveId)
	}

	labels, err := provider.getLabelsForEnclaveObject()
	labels[label_key_consts.KurtosisVolumeTypeLabelKey] = label_value_consts.EnclaveDataVolumeTypeLabelValue

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

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForApiContainer() (KubernetesApiContainerObjectAttributesProvider, error) {
	enclaveId := enclave.EnclaveID(provider.enclaveId)
	return GetKubernetesApiContainerObjectAttributesProvider(enclaveId), nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl)ForUserServiceContainer(serviceID service.ServiceID, serviceGUID service.ServiceGUID, privateIpAddr net.IP, privatePorts map[string]*port_spec.PortSpec) (KubernetesObjectAttributes, error) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForNetworkingSidecarContainer(serviceGUIDSidecarAttachedTo service.ServiceGUID) (KubernetesObjectAttributes, error) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForModuleContainer(
	privateIpAddr net.IP,
	moduleID module.ModuleID,
	moduleGUID module.ModuleGUID,
	privatePortId string,
	privatePortSpec *port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForFilesArtifactExpansionVolume(
	serviceGUID service.ServiceGUID,
	fileArtifactID service.FilesArtifactID,
)(
	KubernetesObjectAttributes,
	error,
){
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForFilesArtifactExpanderContainer(
	guid files_artifact_expander.FilesArtifactExpanderGUID,
)(
	KubernetesObjectAttributes,
	error,
) {
	panic("implement me")
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServiceRegistration (
) (
	KubernetesObjectAttributes,
	error,
) {
	name, err := provider.getNameForEnclaveObject([]string{userServiceRegistrationSuffix})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service registration.")
	}

	labels, err := provider.getLabelsForEnclaveObject()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for user service registration.")
	}
	labels[label_key_consts.KurtosisUserServiceRegistrationTypeLabel] = label_value_consts.UserServiceRegistrationKurtosisResourceTypeLabelValue

	//No userServiceRegistration annotations.
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service registration object attributes.")
	}

	return objectAttributes, nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
// Gets the name for an enclave object, making sure to put the enclave ID first and join using the standardized separator
func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getNameForEnclaveObject(elems []string) (*kubernetes_object_name.KubernetesObjectName, error) {
	toJoin := []string{
		provider.enclaveId,
	}
	toJoin = append(toJoin, elems...)
	nameStr := strings.Join(
		toJoin,
		objectNameElementSeparator,
	)
	name, err := kubernetes_object_name.CreateNewKubernetesObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Kubernetes object name from string '%v'", nameStr)
	}
	return name, nil
}


func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create Kubernetes label value from enclaveId '%v'", provider.enclaveId)
	}
	return map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeLabelKey: label_value_consts.EnclaveKurtosisResourceTypeLabelValue,
		label_key_consts.EnclaveIDLabelKey: enclaveIdLabelValue,
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
	labels[label_key_consts.GUIDLabelKey] = guidLabelValue
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
	labels[label_key_consts.IDLabelKey] = idLabelValue
	return labels, nil
}

func getLabelKeyValuesAsStrings(labels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range labels {
		result[key.GetString()] = value.GetString()
	}
	return result
}

// Gets the name for an artifact expansion object (either volume or container)
func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getArtifactExpansionObjectName(
	objectLabel string,
	forServiceGUID string,
	artifactId string,
) (*kubernetes_object_name.KubernetesObjectName, error) {
	name, err := provider.getNameForEnclaveObject([]string{
		objectLabel,
		"for",
		forServiceGUID,
		"using",
		artifactId,
		"at",
		time.Now().Format(artifactExpansionObjectTimestampFormat), // We add this timestamp so that if the same artifact for the same service GUID expanded twice, we won't get collisions
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the artifact expansion object name")
	}
	return name, nil
}