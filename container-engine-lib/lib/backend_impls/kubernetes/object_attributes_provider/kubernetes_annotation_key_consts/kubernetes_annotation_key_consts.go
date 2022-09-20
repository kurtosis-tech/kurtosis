package kubernetes_annotation_key_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	labelKeyPrefixStr = "kurtosistech.com/"

	portSpecsAnnotationKeyStr = labelKeyPrefixStr + "ports"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var PortSpecsKubernetesAnnotationKey = kubernetes_annotation_key.MustCreateNewKubernetesAnnotationKey(portSpecsAnnotationKeyStr)
