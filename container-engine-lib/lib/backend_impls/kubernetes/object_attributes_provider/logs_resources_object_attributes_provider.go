package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	logsDatabaseNamePrefix = "kurtosis-logs-database"
)

type KubernetesLogsDatabaseObjectAttributesProvider interface {
	ForLogsDatabasePod() (KubernetesObjectAttributes, error)
	ForLogsDatabaseService() (KubernetesObjectAttributes, error)
	ForLogsCollectorsDaemonSet() (KubernetesObjectAttributes, error)
}

type kubernetesLogsDatabaseObjectAttributesProviderImpl struct {
	logsDatabaseGuid logs_database.LogsDatabaseGUID
}

func newKubernetesLogsDatabaseObjectAttributesProviderImpl(logsDatabaseGuid logs_database.LogsDatabaseGUID) *kubernetesLogsDatabaseObjectAttributesProviderImpl {
	return &kubernetesLogsDatabaseObjectAttributesProviderImpl{
		logsDatabaseGuid: logsDatabaseGuid,
	}
}

func GetKubernetesLogsDatabaseObjectAttributesProvider(logsDatabaseGuid logs_database.LogsDatabaseGUID) KubernetesLogsDatabaseObjectAttributesProvider {
	return newKubernetesLogsDatabaseObjectAttributesProviderImpl(logsDatabaseGuid)
}

func (provider *kubernetesLogsDatabaseObjectAttributesProviderImpl) ForLogsDatabasePod() (KubernetesObjectAttributes, error) {
	name, err := provider.getLogsDatabaseObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a pod name for the engine pod")
	}

	labels, err := provider.getLogsDatabaseObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for engine pod
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsDatabaseObjectAttributesProviderImpl) ForLogsDatabaseService() (KubernetesObjectAttributes, error) {
	name, err := provider.getLogsDatabaseObjectName()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a pod name for the engine pod")
	}

	labels, err := provider.getLogsDatabaseObjectLabels()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes labels")
	}

	// No custom annotations for logs database pod
	annotations := map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue{}

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsDatabaseObjectAttributesProviderImpl) ForLogsCollectorsDaemonSet() (KubernetesObjectAttributes, error) {
	//TODO implement me
	panic("implement me")
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func (provider *kubernetesLogsDatabaseObjectAttributesProviderImpl) getLogsDatabaseObjectLabels() (map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue, error) {
	guidLabelValue, err := kubernetes_label_value.CreateNewKubernetesLabelValue(string(provider.logsDatabaseGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing engine GUID '%v' into a Kubernetes label value", provider.logsDatabaseGuid)
	}

	// ID and GUID are the same here
	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		label_key_consts.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsDatabaseKurtosisResourceTypeKubernetesLabelValue,
		label_key_consts.IDKubernetesLabelKey:                   guidLabelValue,
		label_key_consts.GUIDKubernetesLabelKey:                 guidLabelValue,
	}
	return labels, nil
}

func (provider *kubernetesLogsDatabaseObjectAttributesProviderImpl) getLogsDatabaseObjectName() (*kubernetes_object_name.KubernetesObjectName, error) {
	result, err := getCompositeKubernetesObjectName([]string{
		logsDatabaseNamePrefix,
		string(provider.logsDatabaseGuid),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a Kubernetes object name for engine GUID '%v'", provider.logsDatabaseGuid)
	}
	return result, nil
}
