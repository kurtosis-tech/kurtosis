package kubernetes_resource_collectors

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	rbacv1 "k8s.io/api/rbac/v1"
)

// NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE
// Due to not having Go 1.18 generics yet, we have to do all this boilerplate in order to do generic filtering
//  on Kubernetes resources
// This entire file is intended to be copy-pasted if we need to create new CollectMatchingXXXXXX functions
// NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE

// TODO Remove all this when we have Go 1.18 generics
type roleBindingKubernetesResource struct {
	underlying rbacv1.RoleBinding
}
func (resource roleBindingKubernetesResource) getName() string {
	return resource.underlying.Name
}
func (resource roleBindingKubernetesResource) getLabels() map[string]string {
	return resource.underlying.Labels
}
func (resource roleBindingKubernetesResource) getUnderlying() interface{} {
	return resource.underlying
}

// TODO Remove all this when we have Go 1.18 generics
// NOTE: This function is intended to be copy-pasted to create new ones
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
	allObjects, err := kubernetesManager.GetRoleBindingsByLabels(ctx, namespace, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v in namespace '%v'", searchLabels, namespace)
	}
	allKubernetesResources := []kubernetesResource{}
	for _, object := range allObjects.Items {
		allKubernetesResources = append(
			allKubernetesResources,
			roleBindingKubernetesResource{underlying: object},
		)
	}
	filteredKubernetesResources, err := postfilterKubernetesResources(allKubernetesResources, postFilterLabelKey, postFilterLabelValues)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred during postfiltering")
	}
	result := map[string][]*rbacv1.RoleBinding{}
	for labelValue, matchingResources := range filteredKubernetesResources {
		castedObjects := []*rbacv1.RoleBinding{}
		for _, resource := range matchingResources {
			casted, ok := resource.getUnderlying().(rbacv1.RoleBinding)
			if !ok {
				return nil, stacktrace.NewError("An error occurred downcasting Kubernetes resource object '%+v'", resource.getUnderlying())
			}
			castedObjects = append(castedObjects, &casted)
		}
		result[labelValue] = castedObjects
	}
	return result, nil
}
