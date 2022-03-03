package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/label_value_consts"
)

// TODO MOVE TO FOREVER CONSTS!!!!!
var GlobalLabels = map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
	label_key_consts.AppIDLabelKey: label_value_consts.AppIDLabelValue,
	// TODO container engine lib version??
}