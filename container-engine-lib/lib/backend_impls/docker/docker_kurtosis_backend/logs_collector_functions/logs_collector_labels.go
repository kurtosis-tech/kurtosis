package logs_collector_functions

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_labels_for_logs"
)

// These are the list of container's labels that the Docker's logging driver (currently, the Fluentd logging driver)
// will add into the logs stream when it sends them to the destination (currently, Fluentbit logs collector)
type LogsCollectorLabels []string

func GetKurtosisTrackedLogsCollectorLabels() LogsCollectorLabels {

	var logsCollectorLabels LogsCollectorLabels

	allLogsDatabaseKurtosisTrackedDockerLabelsSet := docker_labels_for_logs.GetAllLogsDatabaseKurtosisTrackedDockerLabels()

	for _, dockerLabelKey := range allLogsDatabaseKurtosisTrackedDockerLabelsSet {
		logsCollectorLabel := dockerLabelKey.GetString()
		logsCollectorLabels = append(logsCollectorLabels, logsCollectorLabel)
	}

	return logsCollectorLabels
}
