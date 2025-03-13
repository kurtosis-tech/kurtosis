package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	logsCollectorNamePrefix       = "kurtosis-logs-collector"
	logsCollectorConfigNamePrefix = "kurtosis-logs-collector-config"
)

type KubernetesLogsCollectorObjectAttributesProvider interface {
	ForLogsCollectorDaemonSet() (KubernetesObjectAttributes, error)

	ForLogsCollectorNamespace() (KubernetesObjectAttributes, error)

	ForLogsCollectorConfigMap() (KubernetesObjectAttributes, error)

	ForLogsCollectorServiceAccount() (KubernetesObjectAttributes, error)

	ForLogsCollectorClusterRole() (KubernetesObjectAttributes, error)

	ForLogsCollectorClusterRoleBinding() (KubernetesObjectAttributes, error)
}

func GetKubernetesLogsCollectorObjectAttributesProvider(logsCollectorGuid logs_collector.LogsCollectorGuid) KubernetesLogsCollectorObjectAttributesProvider {
	return newKubernetesLogsCollectorObjectAttributesProvider(logsCollectorGuid)
}

type kubernetesLogsCollectorObjectAttributesProviderImpl struct {
	logsCollectorGuid logs_collector.LogsCollectorGuid
}

func newKubernetesLogsCollectorObjectAttributesProvider(logsCollectorGuid logs_collector.LogsCollectorGuid) *kubernetesLogsCollectorObjectAttributesProviderImpl {
	return &kubernetesLogsCollectorObjectAttributesProviderImpl{
		logsCollectorGuid: logsCollectorGuid,
	}
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorDaemonSet() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, err // already wrapped with propagate
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorConfigMap() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorConfigNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, err // already wrapped with propagate
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorNamespace() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating a Kubernetes object name with prefix '%v' and guid '%v'.", logsCollectorConfigNamePrefix, provider.logsCollectorGuid)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorServiceAccount() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating a Kubernetes object name with prefix '%v' and guid '%v'.", logsCollectorConfigNamePrefix, provider.logsCollectorGuid)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorClusterRole() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating a Kubernetes object name with prefix '%v' and guid '%v'.", logsCollectorConfigNamePrefix, provider.logsCollectorGuid)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsCollectorObjectAttributesProviderImpl) ForLogsCollectorClusterRoleBinding() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsCollectorNamePrefix, string(provider.logsCollectorGuid)})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating a Kubernetes object name with prefix '%v' and guid '%v'.", logsCollectorConfigNamePrefix, provider.logsCollectorGuid)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsCollectorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}
