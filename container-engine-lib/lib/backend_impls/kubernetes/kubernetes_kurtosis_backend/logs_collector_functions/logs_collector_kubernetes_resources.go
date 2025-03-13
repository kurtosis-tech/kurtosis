package logs_collector_functions

import (
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type logsCollectorKubernetesResources struct {
	daemonSet *v1.DaemonSet

	configMap *apiv1.ConfigMap

	namespace *apiv1.Namespace

	// required for retrieving label information for logs from k8s api server
	clusterRole        *rbacv1.ClusterRole
	serviceAccount     *apiv1.ServiceAccount
	clusterRoleBinding *rbacv1.ClusterRoleBinding
}
