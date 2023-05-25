package object_name_constants

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
)

const (
	// The name of the GRPC ports exposed in services
	kurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	kurtosisInternalContainerGrpcProxyPortId = "grpc-proxy"
)

var KurtosisInternalContainerGrpcPortName = kubernetes_object_name.MustCreateNewKubernetesObjectName(kurtosisInternalContainerGrpcPortId)
var KurtosisInternalContainerGrpcProxyPortName = kubernetes_object_name.MustCreateNewKubernetesObjectName(kurtosisInternalContainerGrpcProxyPortId)
