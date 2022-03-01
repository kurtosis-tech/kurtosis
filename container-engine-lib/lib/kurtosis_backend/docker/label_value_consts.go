package docker

const (
	// TODO MOVE TO FOREVER CONSTANTS!!
	appIdLabelValueStr = "kurtosis"

	// TODO MOVE TO FOREVER CONSTANTS!!
	engineContainerTypeLabelValueStr = "kurtosis-engine"






	// The ID of the GRPC port for Kurtosis-internal containers (e.g. API container, engine, modules, etc.) which will
	//  be stored in the port spec label
	kurtosisInternalContainerGrpcPortId = "grpc"

	// The ID of the GRPC proxy port for Kurtosis-internal containers. This is necessary because
	// Typescript's grpc-web cannot communicate directly with GRPC ports, so Kurtosis-internal containers
	// need a proxy  that will translate grpc-web requests before they hit the main GRPC server
	kurtosisInternalContainerGrpcProxyPortId = "grpc-proxy"
)
var AppIDLabelValue = MustCreateNewDockerLabelValue(appIdLabelValueStr)

var EngineContainerTypeLabelValue = MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)
