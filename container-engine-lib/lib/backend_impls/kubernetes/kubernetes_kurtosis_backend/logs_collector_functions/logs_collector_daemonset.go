package logs_collector_functions

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
)

type LogsCollectorDaemonSet interface {
	CreateAndStart(
		ctx context.Context,
		logsDatabaseOrAggregatorHost string,
		logsDatabaseOrAggregatorPort uint16,
		tcpPortNumber uint16,
		httpPortNumber uint16,
		logsCollectorTcpPortId string,
		logsCollectorHttpPortId string,
		objAttrsProvider object_attributes_provider.KubernetesObjectAttributesProvider,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) (string, map[string]string, map[nat.Port]*nat.PortBinding, func(), error)
}
