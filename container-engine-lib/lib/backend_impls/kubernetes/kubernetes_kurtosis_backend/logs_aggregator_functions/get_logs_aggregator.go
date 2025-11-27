package logs_aggregator_functions

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsAggregator(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*logs_aggregator.LogsAggregator, error) {
	obj, _, err := getLogsAggregatorObjAndResourcesForCluster(ctx, kubernetesManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting logs aggregator object for cluster.")
	}
	return obj, nil
}
