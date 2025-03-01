package logs_aggregator_functions

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type LogsAggregatorKubernetesResources struct {
	*apiv1.Service

	*appsv1.Deployment

	*apiv1.Namespace

	// potentially service account, cluster role, cluster role binding
}
