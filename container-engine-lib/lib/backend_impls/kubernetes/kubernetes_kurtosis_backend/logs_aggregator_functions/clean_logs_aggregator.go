package logs_aggregator_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
)

func CleanLogsAggregator(
	ctx context.Context,
	logsAggregatorDeployment LogsAggregatorDeployment,
	kubernetesManager *kubernetes_manager.KubernetesManager) error {
	_, k8sResources, err := getLogsAggregatorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting logs aggregator object and resources for cluster.")
	}

	if err := logsAggregatorDeployment.Clean(ctx, k8sResources.deployment, kubernetesManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred cleaning logs aggregator deployment '%v'", k8sResources.deployment.Name)
	}

	return nil
}
