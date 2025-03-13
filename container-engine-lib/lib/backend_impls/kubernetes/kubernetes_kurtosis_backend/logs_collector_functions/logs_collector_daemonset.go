package logs_collector_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	) (
		*appsv1.DaemonSet,
		*apiv1.ConfigMap,
		*apiv1.Namespace,
		*apiv1.ServiceAccount,
		*rbacv1.ClusterRole,
		*rbacv1.ClusterRoleBinding,
		func(),
		error,
	)

	// GetHttpHealthCheckEndpoint returns endpoint for verifying the availability of the logs collector application on pods managed by the daemon set
	GetHttpHealthCheckEndpoint() string

	// Clean removes any resources the logs collector creates for durability of logs in the case of crashes (e.g. checkpoint dbs)
	Clean(
		ctx context.Context,
		logsCollectorDaemonSet *appsv1.DaemonSet,
		kubernetesManager *kubernetes_manager.KubernetesManager,
	) error
}
