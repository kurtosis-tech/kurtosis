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
	appIdLabelValueStr                      = "kurtosis"
	engineKurtosisResourceTypeLabelValueStr = "kurtosis-engine"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	enclaveKurtosisResourceTypeLabelValueStr           			= "enclave"
	apiContainerKurtosisResourceTypeLabelValueStr      			= "api-container"
	userServiceKurtosisResourceTypeLabelValueStr       			= "user-service"
	networkingSidecarKurtosisResourceTypeLabelValueStr 			= "networking-sidecar"
	moduleKurtosisResourceTypeLabelValueStr            			= "module"
	filesArtifactExpanderKurtosisResourceLabelValueStr 			= "files-artifact-expander"

	enclaveDataVolumeTypeLabelValueStr = "enclave-data"

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
var EngineKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(engineKurtosisResourceTypeLabelValueStr)
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var ModuleKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(moduleKurtosisResourceTypeLabelValueStr)
var EnclaveKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveKurtosisResourceTypeLabelValueStr)
var APIContainerKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(apiContainerKurtosisResourceTypeLabelValueStr)
var UserServiceKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(userServiceKurtosisResourceTypeLabelValueStr)
var FilesArtifactExpanderKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(filesArtifactExpanderKurtosisResourceLabelValueStr)
var NetworkingSidecarKurtosisResourceTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(networkingSidecarKurtosisResourceTypeLabelValueStr)
var EnclaveDataVolumeTypeLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveDataVolumeTypeLabelValueStr)
var NetworkPartitioningEnabledLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(trueValueStr)
var NetworkPartitioningDisabledLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(falseValueStr)