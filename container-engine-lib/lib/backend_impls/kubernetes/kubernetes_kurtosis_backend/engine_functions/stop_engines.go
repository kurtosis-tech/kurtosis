package engine_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
)

func StopEngines(
	ctx context.Context,
	filters *engine.EngineFilters,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	resultSuccessfulEngineGuids map[engine.EngineGUID]bool,
	resultErroredEngineGuids map[engine.EngineGUID]error,
	resultErr error,
) {
	return destroyEngines(ctx, filters, kubernetesManager)
}
