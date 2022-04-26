package object_name_constants

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//

	// The name of the GRPC ports exposed in services
	kurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	kurtosisInternalContainerGrpcProxyPortId = "grpc-proxy"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var KurtosisInternalContainerGrpcPortName = kubernetes_object_name.MustCreateNewKubernetesObjectName(kurtosisInternalContainerGrpcPortId)
var KurtosisInternalContainerGrpcProxyPortName = kubernetes_object_name.MustCreateNewKubernetesObjectName(kurtosisInternalContainerGrpcProxyPortId)
