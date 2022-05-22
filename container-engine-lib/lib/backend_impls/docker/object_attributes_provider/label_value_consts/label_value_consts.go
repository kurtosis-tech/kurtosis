package label_value_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	appIdLabelValueStr               = "kurtosis"
	engineContainerTypeLabelValueStr = "kurtosis-engine"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	apiContainerContainerTypeLabelValueStr          = "api-container"
	userServiceContainerTypeLabelValueStr           = "user-service"
	networkingSidecarContainerTypeLabelValueStr     = "networking-sidecar"
	moduleContainerTypeLabelValueStr                = "module"
	filesArtifactExpanderContainerTypeLabelValueStr = "files-artifact-expander"

	enclaveDataVolumeTypeLabelValueStr            = "enclave-data"
	filesArtifactExpansionVolumeTypeLabelValueStr = "files-artifacts-expansion"

	trueValueStr  = "true"
	falseValueStr = "false"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(appIdLabelValueStr)
var EngineContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)
var ModuleContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(moduleContainerTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var APIContainerContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(apiContainerContainerTypeLabelValueStr)
var UserServiceContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(userServiceContainerTypeLabelValueStr)
var NetworkingSidecarContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(networkingSidecarContainerTypeLabelValueStr)
var NetworkPartitioningEnabledDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(trueValueStr)
var NetworkPartitioningDisabledDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(falseValueStr)
var FilesArtifactExpanderContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactExpanderContainerTypeLabelValueStr)

var EnclaveDataVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(enclaveDataVolumeTypeLabelValueStr)
var FilesArtifactExpansionVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactExpansionVolumeTypeLabelValueStr)