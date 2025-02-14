package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

func GetLogsCollector(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (*logs_collector.LogsCollector, error) {
	//matchingEngines, _, err := getMatchingEngineObjectsAndKubernetesResources(ctx, filters, kubernetesManager)
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	//}
	return nil, nil
}
