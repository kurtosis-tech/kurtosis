package docker_labels_for_logs

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
)

// The following docker labels will be added into the logs stream
// These are necessary for propagating information for log filtering and retrieval through the logging pipeline
var DockerLabelsForLogsStream = []*docker_label_key.DockerLabelKey{
	docker_label_key.ContainerTypeDockerLabelKey,
	docker_label_key.LogsEnclaveUUIDDockerLabelKey,
	docker_label_key.LogsServiceUUIDDockerLabelKey,
	docker_label_key.LogsServiceShortUUIDDockerLabelKey,
	docker_label_key.LogsServiceNameDockerLabelKey,
}

func GetDockerLabelsForLogStream() []*docker_label_key.DockerLabelKey {
	return DockerLabelsForLogsStream
}
