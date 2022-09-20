package kubernetes_resource_collectors

import "github.com/kurtosis-tech/stacktrace"

// Finds namespaces using labels, postfilters them using a label value, and returns them categorized by that label
func postfilterKubernetesResources(
	resources []kubernetesResource,
	postFilterLabelKey string,
	// A nil or empty map will match all values
	postFilterLabelValues map[string]bool,
) (
	map[string][]kubernetesResource,
	error,
) {
	result := map[string][]kubernetesResource{}

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

		matchingResources, found := result[labelValue]
		if !found {
			matchingResources = []kubernetesResource{}
		}
		result[labelValue] = append(matchingResources, resource)
	}
	return result, nil
}
