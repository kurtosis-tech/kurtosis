package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	logsAggregatorNamePrefix       = "kurtosis-logs-aggregator"
	logsAggregatorConfigNamePrefix = "kurtosis-logs-aggregator-config"
)

type KubernetesLogsAggregatorObjectAttributesProvider interface {
	ForLogsAggregatorDeployment() (KubernetesObjectAttributes, error)

	ForLogsAggregatorNamespace() (KubernetesObjectAttributes, error)

	ForLogsAggregatorService() (KubernetesObjectAttributes, error)
}

func GetKubernetesLogsAggregatorObjectAttributesProvider(logsAggregatorGuid logs_aggregator.LogsAggregatorGuid) KubernetesLogsAggregatorObjectAttributesProvider {
	return newKubernetesLogsAggregatorObjectAttributesProvider(logsAggregatorGuid)
}

type kubernetesLogsAggregatorObjectAttributesProviderImpl struct {
	logsAggregatorGuid logs_aggregator.LogsAggregatorGuid
}

func newKubernetesLogsAggregatorObjectAttributesProvider(logsAggregatorGuid logs_aggregator.LogsAggregatorGuid) *kubernetesLogsAggregatorObjectAttributesProviderImpl {
	return &kubernetesLogsAggregatorObjectAttributesProviderImpl{
		logsAggregatorGuid: logsAggregatorGuid,
	}
}

func (provider *kubernetesLogsAggregatorObjectAttributesProviderImpl) ForLogsAggregatorDeployment() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsAggregatorNamePrefix, string(provider.logsAggregatorGuid)})
	if err != nil {
		return nil, err // already wrapped with propagate
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsAggregatorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsAggregatorObjectAttributesProviderImpl) ForLogsAggregatorService() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsAggregatorConfigNamePrefix, string(provider.logsAggregatorGuid)})
	if err != nil {
		return nil, err // already wrapped with propagate
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsAggregatorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}

func (provider *kubernetesLogsAggregatorObjectAttributesProviderImpl) ForLogsAggregatorNamespace() (KubernetesObjectAttributes, error) {
	name, err := getCompositeKubernetesObjectName([]string{logsAggregatorNamePrefix, string(provider.logsAggregatorGuid)})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating a Kubernetes object name with prefix '%v' and guid '%v'.", logsAggregatorConfigNamePrefix, provider.logsAggregatorGuid)
	}

	labels := map[*kubernetes_label_key.KubernetesLabelKey]*kubernetes_label_value.KubernetesLabelValue{
		kubernetes_label_key.KurtosisResourceTypeKubernetesLabelKey: label_value_consts.LogsAggregatorKurtosisResourceTypeKubernetesLabelValue,
	}

	annotations := make(map[*kubernetes_annotation_key.KubernetesAnnotationKey]*kubernetes_annotation_value.KubernetesAnnotationValue)

	objectAttributes, err := newKubernetesObjectAttributesImpl(name, labels, annotations)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the Kubernetes object attributes with the name "+
			"'%s' and labels '%+v', and annotations '%+v'", name.GetString(), labels, annotations)
	}
	return objectAttributes, nil
}
