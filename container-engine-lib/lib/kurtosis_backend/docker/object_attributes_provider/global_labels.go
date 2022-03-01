package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_value"
)

// TODO MOVE TO FOREVER CONSTS!!!!!
var GlobalLabels = map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
	AppIDLabelKey: AppIDLabelValue,
	// TODO container engine lib version??
}