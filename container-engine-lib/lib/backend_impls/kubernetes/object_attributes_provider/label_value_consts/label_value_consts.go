package label_value_consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_value"
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

	enclaveKurtosisResourceTypeLabelValueStr      = "enclave"
	apiContainerKurtosisResourceTypeLabelValueStr = "api-container"
	userServiceKurtosisResourceTypeLabelValueStr  = "user-service"

	enclaveDataVolumeTypeLabelValueStr             = "enclave-data"
	filesArtifactsExpansionVolumeTypeLabelValueStr = "files-artifacts-expansion"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
// which will cause a resource leak on the user's system!
//
// If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var AppIDKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(appIdLabelValueStr)
var EngineKurtosisResourceTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(engineKurtosisResourceTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var EnclaveKurtosisResourceTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveKurtosisResourceTypeLabelValueStr)
var APIContainerKurtosisResourceTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(apiContainerKurtosisResourceTypeLabelValueStr)
var UserServiceKurtosisResourceTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(userServiceKurtosisResourceTypeLabelValueStr)
var EnclaveDataVolumeTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(enclaveDataVolumeTypeLabelValueStr)
var FilesArtifactsExpansionVolumeTypeKubernetesLabelValue = kubernetes_label_value.MustCreateNewKubernetesLabelValue(filesArtifactsExpansionVolumeTypeLabelValueStr)
