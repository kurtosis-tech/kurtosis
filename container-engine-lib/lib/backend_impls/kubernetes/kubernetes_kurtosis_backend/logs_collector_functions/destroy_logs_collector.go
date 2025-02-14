package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
)

// DestroyLogsCollector Destroys the logs collector and its associated resources
func DestroyLogsCollector(ctx context.Context, kubernetesManager *kubernetes_manager.KubernetesManager) error {

	return nil
}
