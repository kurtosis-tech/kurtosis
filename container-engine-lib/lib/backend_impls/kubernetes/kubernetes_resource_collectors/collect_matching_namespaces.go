package kubernetes_resource_collectors

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	apiv1 "k8s.io/api/core/v1"
)

// NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE
// Due to not having Go 1.18 generics yet, we have to do all this boilerplate in order to do generic filtering
//  on Kubernetes resources
// This entire file is intended to be copy-pasted if we need to create new CollectMatchingXXXXXX functions
// NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE NOTE

// TODO Remove all this when we have Go 1.18 generics
type namespaceKubernetesResource struct {
	underlying apiv1.Namespace
}
func (resource namespaceKubernetesResource) getName() string {
	return resource.underlying.Name
}
func (resource namespaceKubernetesResource) getLabels() map[string]string {
	return resource.underlying.Labels
}
func (resource namespaceKubernetesResource) getUnderlying() interface{} {
	return resource.underlying
}

// TODO Remove all this when we have Go 1.18 generics
// NOTE: This function is intended to be copy-pasted to create new ones
func CollectMatchingNamespaces(
	ctx context.Context,
	kubernetesManager *kubernetes_manager.KubernetesManager,
	searchLabels map[string]string,
	postFilterLabelKey string,
	postFilterLabelValues map[string]bool,
) (
	map[string]*apiv1.Namespace,
	error,
) {
	allObjects, err := kubernetesManager.GetNamespacesByLabels(ctx, searchLabels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes resources matching labels: %+v", searchLabels)
	}
	allKubernetesResources := []kubernetesResource{}
	for _, object := range allObjects.Items {
		allKubernetesResources = append(
			allKubernetesResources,
			namespaceKubernetesResource{underlying: object},
		)
	}
	filteredKubernetesResources, err := postfilterKubernetesResources(allKubernetesResources, postFilterLabelKey, postFilterLabelValues)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred during postfiltering")
	}
	result := map[string]*apiv1.Namespace{}
	for labelValue, uncastedResource := range filteredKubernetesResources {
		castedResource, ok := uncastedResource.getUnderlying().(apiv1.Namespace)
		if !ok {
			return nil, stacktrace.NewError("An error occurred downcasting Kubernetes resource object '%+v'", uncastedResource.getUnderlying())
		}
		result[labelValue] = &castedResource
	}
	return result, nil
}

