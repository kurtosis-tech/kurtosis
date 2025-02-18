package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsCollector(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*logs_collector.LogsCollector, error) {
	obj, _, err := getLogsCollectorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs collector object for cluster.")
	}
	return obj, nil
}
