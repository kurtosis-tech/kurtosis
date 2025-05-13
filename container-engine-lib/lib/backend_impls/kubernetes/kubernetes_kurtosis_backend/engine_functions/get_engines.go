package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/stacktrace"
)

func GetEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	engineNodeName string,
) (map[engine.EngineGUID]*engine.Engine, error) {
	matchingEngines, _, err := getMatchingEngineObjectsAndKubernetesResources(ctx, filters, kubernetesManager, engineNodeName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines matching the following filters: %+v", filters)
	}
	return matchingEngines, nil
}
