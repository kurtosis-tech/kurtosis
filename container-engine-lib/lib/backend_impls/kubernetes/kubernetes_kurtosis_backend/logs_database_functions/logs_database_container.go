package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
)

type LogsDatabaseContainer interface {
	CreateAndStart(
		ctx context.Context,
		engineNamespace string,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) error
}
