package logs_aggregator_functions

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type logsAggregatorKubernetesResources struct {
	service *apiv1.Service

	deployment *appsv1.Deployment

	namespace *apiv1.Namespace

	// potentially service account, cluster role, cluster role binding
}
