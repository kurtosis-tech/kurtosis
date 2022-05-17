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

	enclaveDataVolumeTypeLabelValueStr = "enclave-data"
	filesArtifactExpansionVolumeTypeLabelValueStr = "files-artifact-expansion"

	trueValueStr  = "true"
	falseValueStr = "false"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(appIdLabelValueStr)
var EngineContainerTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)
var ModuleContainerTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(moduleContainerTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var APIContainerContainerTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(apiContainerContainerTypeLabelValueStr)
var UserServiceContainerTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(userServiceContainerTypeLabelValueStr)
var NetworkingSidecarContainerTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(networkingSidecarContainerTypeLabelValueStr)
var NetworkPartitioningEnabledKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(trueValueStr)
var NetworkPartitioningDisabledKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(falseValueStr)
var FilesArtifactExpanderContainerTypeKuberenetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactExpanderContainerTypeLabelValueStr)

var EnclaveDataVolumeTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(enclaveDataVolumeTypeLabelValueStr)
var FilesArtifactExpansionVolumeTypeKubernetesLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactExpansionVolumeTypeLabelValueStr)