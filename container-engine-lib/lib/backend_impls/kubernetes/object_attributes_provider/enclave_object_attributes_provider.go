package object_attributes_provider

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	namespacePrefix = "kt"

	enclaveDataDirFragment = "enclave-data-dir"

	traefikIngressRouterEntrypointsValue = "web"
)

type KubernetesEnclaveObjectAttributesProvider interface {
	ForEnclaveNamespace(creationTime time.Time, enclaveName string) (KubernetesObjectAttributes, error)
	ForApiContainer() KubernetesApiContainerObjectAttributesProvider
	ForEnclaveDataDirVolume() (KubernetesObjectAttributes, error)
	ForUserServiceService(
		uuid service.ServiceUUID,
		id service.ServiceName,
	) (KubernetesObjectAttributes, error)
	ForUserServicePod(
		uuid service.ServiceUUID,
		id service.ServiceName,
		privatePorts map[string]*port_spec.PortSpec,
		userLabels map[string]string,
	) (KubernetesObjectAttributes, error)
	ForSinglePersistentDirectoryVolume(
		persistentKey service_directory.DirectoryPersistentKey,
	) (KubernetesObjectAttributes, error)
	ForUserServiceIngress(
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

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveNamespace(creationTime time.Time, enclaveName string) (KubernetesObjectAttributes, error) {
	// TODO: might need to revert this if we have multiple users on the same cluster (what if two people create enclaves with name test?)
	name, err := getCompositeKubernetesObjectName([]string{
		namespacePrefix,
		enclaveName,
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

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServiceService(
	serviceUUID service.ServiceUUID,
	serviceName service.ServiceName,
) (KubernetesObjectAttributes, error) {
	name, err := getKubernetesObjectName(serviceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service service.")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(serviceName), string(serviceUUID))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service service with name '%s' and UUID '%s'.",
			string(serviceName),
			string(serviceUUID),
		)
	}
	labels[kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	//No userServiceService annotations.
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service service object attributes.")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServicePod(
	serviceUUID service.ServiceUUID,
	serviceName service.ServiceName,
	privatePorts map[string]*port_spec.PortSpec,
	userLabels map[string]string,
) (KubernetesObjectAttributes, error) {
	name, err := getKubernetesObjectName(serviceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service pod")
	}

	serializedPortSpecsAnnotationValue, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following user service port specs to a string for storing in the ports label: %+v", privatePorts)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(serviceName), string(serviceUUID))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service pod with name '%s' and UUID '%s'",
			serviceName,
			serviceUUID,
		)
	}
	labels[kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	// add user custom label
	for userLabelKey, userLabelValue := range userLabels {
		kubernetesLabelKey, err := kubernetes_label_key.CreateNewKubernetesUserCustomLabelKey(userLabelKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new user custom Kubernetes label key '%s'", userLabelKey)
		}
		kubernetesLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(userLabelValue)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new user custom Kubernetes label value '%s'", userLabelValue)
		}
		labels[kubernetesLabelKey] = kubernetesLabelValue
	}

	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey: serializedPortSpecsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service pod object attributes")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForEnclaveDataDirVolume() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{
		enclaveDataDirFragment,
		provider.enclaveId,
	})

	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", provider.enclaveId)
	}

	hasher := md5.New()
	hasher.Write([]byte(provider.enclaveId))
	hasher.Write([]byte(enclaveDataDirFragment))
	volumeHash := hex.EncodeToString(hasher.Sum(nil))

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(
		enclaveDataDirFragment,
		volumeHash,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get labels for enclave namespace using ID '%v'", provider.enclaveId)
	}

	//No userServiceService annotations.
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service persistent directory object attributes")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForSinglePersistentDirectoryVolume(persistentKey service_directory.DirectoryPersistentKey) (KubernetesObjectAttributes, error) {
	hasher := md5.New()
	hasher.Write([]byte(provider.enclaveId))
	hasher.Write([]byte(persistentKey))
	persistentKeyHash := hex.EncodeToString(hasher.Sum(nil))

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(persistentKey), persistentKeyHash)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for persistent volume with key '%s' and UUID '%s'",
			persistentKey,
			persistentKeyHash,
		)
	}

	//No userServiceService annotations.
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	name, err := getKubernetesPersistentDirectoryName(string(persistentKey))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service persistent directory name for hash: '%s'", persistentKeyHash)
	}
	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service persistent directory object attributes")
	}

	return objectAttributes, nil
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) ForUserServiceIngress(
	serviceUUID service.ServiceUUID,
	serviceName service.ServiceName,
	privatePorts map[string]*port_spec.PortSpec,
) (KubernetesObjectAttributes, error) {
	name, err := getKubernetesObjectName(serviceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get name for user service ingress")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(string(serviceName), string(serviceUUID))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to get labels for user service ingress with name '%s' and UUID '%s'",
			serviceName,
			serviceUUID,
		)
	}
	labels[kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey] = label_value_consts.UserServiceKurtosisResourceTypeKubernetesLabelValue

	traefikIngressRouterEntrypointsAnnotationValue, err := kubernetes_annotation_value.CreateNewKubernetesAnnotationValue(traefikIngressRouterEntrypointsValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new user custom Kubernetes label value '%s'", traefikIngressRouterEntrypointsValue)
	}
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{
		kubernetes_annotation_key_consts.TraefikIngressRouterEntrypointsAnnotationKey: traefikIngressRouterEntrypointsAnnotationValue,
	}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create user service ingress object attributes")
	}

	return objectAttributes, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func getKubernetesObjectName(
	serviceName service.ServiceName,
) (*kubernetes_object_name.KubernetesObjectName, error) {
	name, err := getCompositeKubernetesObjectName(
		[]string{
			string(serviceName),
		})
	return name, err
}

func getKubernetesPersistentDirectoryName(
	persistentKey string,
) (*kubernetes_object_name.KubernetesObjectName, error) {
	name, err := getCompositeKubernetesObjectName(
		[]string{
			persistentKey,
		})
	return name, err
}

func (provider *kubernetesEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	enclaveIdLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(provider.enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create Kubernetes label value from enclaveId '%v'", provider.enclaveId)
	}
	return map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.EnclaveKurtosisResourceTypeKubernetesLabelValue,
		kubernetes_label_key.EnclaveUUIDKubernetesLabelKey:          enclaveIdLabelValue,
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
	labels[kubernetes_label_key.GUIDKubernetesLabelKey] = uuidLabelValue
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
	labels[kubernetes_label_key.IDKubernetesLabelKey] = idLabelValue
	return labels, nil
}

func getLabelKeyValuesAsStrings(labels map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range labels {
		result[key.GetString()] = value.GetString()
	}
	return result
}
