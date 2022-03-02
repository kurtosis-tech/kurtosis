package object_attributes_provider

import "github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_value"



const (
	// TODO MOVE TO FOREVER CONSTANTS!!
	appIdLabelValueStr = "kurtosis"

	// TODO MOVE TO FOREVER CONSTANTS!!
	engineContainerTypeLabelValueStr = "kurtosis-engine"






)
var AppIDLabelValue = docker_label_value.MustCreateNewDockerLabelValue(appIdLabelValueStr)

var EngineContainerTypeLabelValue = docker_label_value.MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)
