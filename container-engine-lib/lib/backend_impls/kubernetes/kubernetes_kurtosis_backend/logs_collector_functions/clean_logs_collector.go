package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
)

func CleanLogsCollector(
	ctx context.Context,
	logsCollectorDaemonSet LogsCollectorDaemonSet,
	kubernetesManager *kubernetes_manager.KubernetesManager) error {
	_, k8sResources, err := getLogsCollectorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting logs collector object and resources for cluster.")
	}

	if err := logsCollectorDaemonSet.Clean(ctx, k8sResources.daemonSet, kubernetesManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred cleaning logs collector daemon set '%v'", k8sResources.daemonSet.Name)
	}

	return nil
}
