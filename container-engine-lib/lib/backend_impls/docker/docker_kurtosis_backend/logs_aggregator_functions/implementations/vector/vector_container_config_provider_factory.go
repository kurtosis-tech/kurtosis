package vector

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

func createVectorContainerConfigProvider(
	portNumber uint16,
	httpPortNumber uint16,
	sinks logs_aggregator.Sinks,
) *vectorContainerConfigProvider {
	config := newVectorConfig(portNumber, httpPortNumber, sinks)
	return newVectorContainerConfigProvider(config, httpPortNumber)
}
