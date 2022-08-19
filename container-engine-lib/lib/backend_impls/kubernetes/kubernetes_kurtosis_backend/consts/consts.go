package consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"k8s.io/api/core/v1"
)

const (
	// The Kurtosis servers (Engine and API Container) use gRPC so MUST listen on TCP (no other protocols are supported), which also
	// means that its grpc-proxy must listen on TCP
	KurtosisServersPortProtocol = port_spec.PortProtocol_TCP

	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	KurtosisInternalContainerGrpcPortSpecId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	KurtosisInternalContainerGrpcProxyPortSpecId = "grpc-proxy"
)

// This maps a Kubernetes pod's phase to a binary "is the pod considered running?" determiner
// Its completeness is enforced via unit test
var IsPodRunningDeterminer = map[v1.PodPhase]bool{
	v1.PodPending:   true,
	v1.PodRunning:   true,
	v1.PodSucceeded: false,
	v1.PodFailed:    false,
	v1.PodUnknown:   false, //We cannot say that a pod is not running if we don't know the real state
}

// Completeness enforced via unit test
var KurtosisPortProtocolToKubernetesPortProtocolTranslator = map[port_spec.PortProtocol]v1.Protocol{
	port_spec.PortProtocol_TCP:  v1.ProtocolTCP,
	port_spec.PortProtocol_UDP:  v1.ProtocolUDP,
	port_spec.PortProtocol_SCTP: v1.ProtocolSCTP,
}
