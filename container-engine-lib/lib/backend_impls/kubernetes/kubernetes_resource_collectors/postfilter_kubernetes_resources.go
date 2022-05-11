package kubernetes_resource_collectors

import "github.com/kurtosis-tech/stacktrace"

// Takes in a list of Kubernetes objects, filters them by the given label, and returns a list of matching ones
// An error is thrown if more than one resource matches the same label value
func postfilterKubernetesResources(
	resources []kubernetesResource,
	postFilterLabelKey string,
// A nil or empty map will match all values
	postFilterLabelValues map[string]bool,
) (
	map[string]kubernetesResource,
	error,
) {
	result := map[string]kubernetesResource{}

	for _, resource := range resources {
		labelValue, hasLabel := resource.getLabels()[postFilterLabelKey]
		if !hasLabel {
			return nil, stacktrace.NewError(
				"Expected to find label '%v' on Kubernetes resource with name '%v' but none was found",
				postFilterLabelKey,
				resource.getName(),
			)
		}

		if postFilterLabelValues != nil && len(postFilterLabelValues) > 0 {
			if _, found := postFilterLabelValues[labelValue]; !found {
				continue
			}
		}

		// We don't want to tolerate multiple resources that have the exact same label value
		preexistingResource, found := result[labelValue]
		if found {
			return nil, stacktrace.NewError(
				"Encountered Kubernetes resource with name '%v', label '%v', and label value '%v' that collides with already-seen " +
					"resource with name '%v'",
				resource.getName(),
				postFilterLabelKey,
				labelValue,
				preexistingResource.getName(),
			)
		}

		result[labelValue] =  resource
	}
	return result, nil
}
