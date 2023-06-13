package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
)

type LogsCollectorContainer interface {
	CreateAndStart(
		ctx context.Context,
		engineNamespace string,
		logsDatabaseHost string,
		logsDatabasePort uint16,
		tcpPortNumber uint16,
		httpPortNumber uint16,
		serviceAccountName string,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) error

	AlreadyExists(
		ctx context.Context,
		engineNamespace string,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) (LogsCollectorContainer, error)
}
