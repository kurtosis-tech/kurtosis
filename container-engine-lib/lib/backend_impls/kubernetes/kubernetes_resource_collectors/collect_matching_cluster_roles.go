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
type clusterRoleKubernetesResource struct {
	underlying rbacv1.ClusterRole
}

func (resource clusterRoleKubernetesResource) getName() string {
	return resource.underlying.Name
}
func (resource clusterRoleKubernetesResource) getLabels() map[string]string {
	return resource.underlying.Labels
}
func (resource clusterRoleKubernetesResource) getUnderlying() interface{} {
	return resource.underlying
}

// TODO Remove all this when we have Go 1.18 generics
// NOTE: This function is intended to be copy-pasted to create new ones
func CollectMatchingClusterRoles(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string][]*rbacv1.ClusterRole,
	error,
) {
	allObjects, err := kubernetesManager.GetClusterRolesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	allKubernetesResources := []kubernetesResource{}
	for _, object := range allObjects.Items {
		allKubernetesResources = append(
			allKubernetesResources,
			clusterRoleKubernetesResource{underlying: object},
		)
	}
	filteredKubernetesResources, err := postfilterKubernetesResources(allKubernetesResources, postFilterLabelKey, postFilterLabelValues)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred during postfiltering")
	}
	result := map[string][]*rbacv1.ClusterRole{}
	for labelValue, matchingResources := range filteredKubernetesResources {
		castedObjects := []*rbacv1.ClusterRole{}
		for _, resource := range matchingResources {
			casted, ok := resource.getUnderlying().(rbacv1.ClusterRole)
			if !ok {
				return nil, stacktrace.NewError("An error occurred downcasting Kubernetes resource object '%+v'", resource.getUnderlying())
			}
			castedObjects = append(castedObjects, &casted)
		}
		result[labelValue] = castedObjects
	}
	return result, nil
}
