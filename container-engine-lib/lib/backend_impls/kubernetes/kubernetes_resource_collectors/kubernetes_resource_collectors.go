package kubernetes_resource_collectors

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func CollectMatchingServiceAccounts(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*apiv1.ServiceAccount,
	error,
) {
	objects, err := kubernetesManager.GetServiceAccountsByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingNamespaces(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*apiv1.Namespace,
	error,
) {
	objects, err := kubernetesManager.GetNamespacesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingPods(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*apiv1.Pod,
	error,
) {
	objects, err := kubernetesManager.GetPodsByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingRoleBindings(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*rbacv1.RoleBinding,
	error,
) {
	objects, err := kubernetesManager.GetRoleBindingsByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingRoles(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*rbacv1.Role,
	error,
) {
	objects, err := kubernetesManager.GetRolesByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingServices(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*apiv1.Service,
	error,
) {
	objects, err := kubernetesManager.GetServicesByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingIngresses(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	namespace string,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*netv1.Ingress,
	error,
) {
	objects, err := kubernetesManager.GetIngressesByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingClusterRoles(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool) (map[string][]*rbacv1.ClusterRole, error) {
	objects, err := kubernetesManager.GetClusterRolesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func CollectMatchingClusterRoleBindings(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*rbacv1.ClusterRoleBinding,
	error,
) {
	objects, err := kubernetesManager.GetClusterRoleBindingsByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	return postFilterKubernetesResources(getListOfPointersFromListOfElements(objects.Items), postFilterLabelKey, postFilterLabelValues)
}

func getListOfPointersFromListOfElements[T any](list []T) []*T {
	newList := []*T{}
	for idx := range list {
		newList = append(newList, &list[idx])
	}
	return newList
}

type kubernetesResource interface {
	GetName() string
	GetLabels() map[string]string
}

// Finds namespaces using labels, postfilters them using a label value, and returns them categorized by that label
func postFilterKubernetesResources[T kubernetesResource](
	resources []T,
	postFilterLabelKey string,
	// A nil or empty map will match all values
	postFilterLabelValues map[string]bool,
) (
	map[string][]T,
	error,
) {
	result := map[string][]T{}
	for _, resource := range resources {
		labelValue, hasLabel := resource.GetLabels()[postFilterLabelKey]
		if !hasLabel {
			return nil, stacktrace.NewError(
				"Expected to find label '%v' on Kubernetes resource with name '%v' but none was found",
				postFilterLabelKey,
				resource.GetName(),
			)
		}

		if len(postFilterLabelValues) > 0 {
			if _, found := postFilterLabelValues[labelValue]; !found {
				continue
			}
		}

		matchingResources, found := result[labelValue]
		if !found {
			matchingResources = []T{}
		}
		result[labelValue] = append(matchingResources, resource)
	}
	return result, nil
}
