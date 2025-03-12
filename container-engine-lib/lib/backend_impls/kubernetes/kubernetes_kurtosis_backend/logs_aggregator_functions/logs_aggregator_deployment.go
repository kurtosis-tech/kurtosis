package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

type LogsAggregatorDeployment interface {
	CreateAndStart(
		ctx context.Context,
		// This is the port that this LogsAggregatorDaemonSet will listen for logs on
		// LogsCollectors should forward logs to this port
		logsListeningPort uint16,
		// Provided so deployment can be scheduled on same node as engine
		engineNamespace string,
		objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) (
		*apiv1.Service,
		*appsv1.Deployment,
		*apiv1.Namespace,
		*apiv1.ConfigMap,
		func(),
		error)

	// GetLogsBaseDirPath returns a string of the base directory path logs are output to on pods associated with the deployment
	GetLogsBaseDirPath() string

	// GetHTTPHealthCheckEndpointAndPort returns a string and int of the http endpoint and port to request aggregator health status
	GetHTTPHealthCheckEndpointAndPort() (string, uint16)

	// Clean removes any resources the logs aggregator creates for durability of logs in the case of crashes (e.g. disk buffers)
	Clean(
		ctx context.Context,
		logsAggregator *appsv1.Deployment,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) error
}
