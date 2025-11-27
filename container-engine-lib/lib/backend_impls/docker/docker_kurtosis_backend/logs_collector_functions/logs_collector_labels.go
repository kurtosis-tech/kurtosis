package logs_collector_functions

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_labels_for_logs"
)

// GetServiceLabelsForLogsTracking returns list of labels to add to kurtosis user service containers
// Docker's logging driver (currently, the Fluentd logging driver) will add them into the logs stream forwarded by the logs collector
func GetLabelsForLogsTrackingLogsOfUserServiceContainers() []string {
	var labels []string

	dockerLabelsForLogsStream := docker_labels_for_logs.GetDockerLabelsForLogStream()

	for _, dockerLabelKey := range dockerLabelsForLogsStream {
		logsCollectorLabel := dockerLabelKey.GetString()
		labels = append(labels, logsCollectorLabel)
	}

	return labels
}
