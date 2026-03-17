package logs_aggregator_functions

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type logsAggregatorKubernetesResources struct {
	service *apiv1.Service

	deployment *appsv1.Deployment

	configMap *apiv1.ConfigMap

	namespace *apiv1.Namespace
}
