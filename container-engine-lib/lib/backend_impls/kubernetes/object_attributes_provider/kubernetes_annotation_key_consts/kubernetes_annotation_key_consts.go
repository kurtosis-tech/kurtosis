package kubernetes_annotation_key_consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	labelKeyPrefixStr = "kurtosistech.com/"

	portSpecsAnnotationKeyStr = labelKeyPrefixStr + "ports"

	enclaveCreationTimeKeyStr = labelKeyPrefixStr + "enclave-creation-time"

	enclaveNameKeyStr = labelKeyPrefixStr + "enclave-name"

	// Traefik ingress router
	traefikKeyIngressRouterPrefixStr = "traefik.ingress.kubernetes.io/router."
	traefikKeyEntrypointsStr         = traefikKeyIngressRouterPrefixStr + "entrypoints"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//
//	If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var PortSpecsKubernetesAnnotationKey = kubernetes_annotation_key.MustCreateNewKubernetesAnnotationKey(portSpecsAnnotationKeyStr)
var EnclaveCreationTimeAnnotationKey = kubernetes_annotation_key.MustCreateNewKubernetesAnnotationKey(enclaveCreationTimeKeyStr)
var EnclaveNameAnnotationKey = kubernetes_annotation_key.MustCreateNewKubernetesAnnotationKey(enclaveNameKeyStr)
var TraefikIngressRouterEntrypointsAnnotationKey = kubernetes_annotation_key.MustCreateNewKubernetesAnnotationKey(traefikKeyEntrypointsStr)
