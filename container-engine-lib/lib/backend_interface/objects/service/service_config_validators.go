package service

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
	"github.com/kurtosis-tech/stacktrace"
)

func ValidateServiceConfigLabels(labels map[string]string) error {
	for labelKey, labelValue := range labels {

		// key validations
		if err := docker_label_key.ValidateUserCustomLabelKey(labelKey); err != nil {
			return stacktrace.Propagate(err, "Invalid service config label key '%s'", labelKey)
		}
		if err := kubernetes_label_key.ValidateUserCustomLabelKey(labelKey); err != nil {
			return stacktrace.Propagate(err, "Invalid service config label key '%s'", labelKey)
		}

		// values validations
		if err := docker_label_value.ValidateDockerLabelValue(labelValue); err != nil {
			return stacktrace.Propagate(err, "Invalid service config label value '%s'", labelValue)
		}
		if err := kubernetes_label_value.ValidateKubernetesLabelValue(labelValue); err != nil {
			return stacktrace.Propagate(err, "Invalid service config label value '%s'", labelValue)
		}
	}
	return nil
}
