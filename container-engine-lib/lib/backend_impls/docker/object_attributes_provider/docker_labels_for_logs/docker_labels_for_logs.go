package docker_labels_for_logs

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
)

// The following docker labels will be added into the logs stream
// These are necessary for propagating information for log filtering and retrieval through the logging pipeline
var LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream = []*docker_label_key.DockerLabelKey{
	label_key_consts.ContainerTypeDockerLabelKey,
	label_key_consts.UserServiceGUIDDockerLabelKey,
	label_key_consts.LogsEnclaveIDDockerLabelKey,
	label_key_consts.EnclaveUUIDDockerLabelKey,
	label_key_consts.EnclaveNameDockerLabelKey,
}

// These are all the logs database Kurtosis tracked Docker Labels used
func GetAllLogsDatabaseKurtosisTrackedDockerLabels() []*docker_label_key.DockerLabelKey {
	return LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream
}
