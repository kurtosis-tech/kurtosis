package label_value_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	appIdLabelValueStr              = "kurtosis"
	engineResourceTypeLabelValueStr = "kurtosis-engine"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	enclaveResourceTypeLabelValueStr                = "enclave"
	apiContainerContainerTypeLabelValueStr          = "api-container"
	userServiceContainerTypeLabelValueStr           = "user-service"
	networkingSidecarContainerTypeLabelValueStr     = "networking-sidecar"
	moduleContainerTypeLabelValueStr                = "module"
	filesArtifactExpanderContainerTypeLabelValueStr = "files-artifact-expander"
	enclaveDataVolumeTypeLabelValueStr              = "enclave-data"

	trueValueStr  = "true"
	falseValueStr = "false"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(appIdLabelValueStr)
var EngineResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(engineResourceTypeLabelValueStr)
var ModuleContainerTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(moduleContainerTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
var EnclaveResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveResourceTypeLabelValueStr)
var APIContainerContainerTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(apiContainerContainerTypeLabelValueStr)
var UserServiceContainerTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(userServiceContainerTypeLabelValueStr)
var NetworkingSidecarContainerTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(networkingSidecarContainerTypeLabelValueStr)
var NetworkPartitioningEnabledLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(trueValueStr)
var NetworkPartitioningDisabledLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(falseValueStr)
var FilesArtifactExpanderContainerTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(filesArtifactExpanderContainerTypeLabelValueStr)
var EnclaveDataVolumeTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveDataVolumeTypeLabelValueStr)