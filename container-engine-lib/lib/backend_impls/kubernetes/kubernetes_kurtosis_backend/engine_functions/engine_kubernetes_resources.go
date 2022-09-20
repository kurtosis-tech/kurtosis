package engine_functions

import (
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Any of these values being nil indicates that the resource doesn't exist
type engineKubernetesResources struct {
	clusterRole *rbacv1.ClusterRole

	clusterRoleBinding *rbacv1.ClusterRoleBinding

	namespace *apiv1.Namespace

	// Should always be nil if namespace is nil
	serviceAccount *apiv1.ServiceAccount

	// Should always be nil if namespace is nil
	service *apiv1.Service

	// Should always be nil if namespace is nil
	pod *apiv1.Pod
}
