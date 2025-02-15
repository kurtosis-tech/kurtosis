package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO: implement
// DestroyLogsCollector Destroys the logs collector and its associated resources
func DestroyLogsCollector(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) error {
	logsCollectorResources, err := getLogsCollectorKubernetesResourcesForCluster(ctx, "kube-system", kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving kubernetes resources for logs collector.")
	}

	if logsCollectorResources.daemonSet != nil {
		if err := kubernetesManager.RemoveDaemonSet(ctx, "kube-system", logsCollectorResources.daemonSet); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing logs collector daemon set.")
		}
	}

	if logsCollectorResources.configMap != nil {
		if err := kubernetesManager.RemoveConfigMap(ctx, "kube-system", logsCollectorResources.configMap); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing logs collector config map.")
		}
	}

	return nil
}
