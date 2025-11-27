package vector

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

func createVectorConfigurationCreatorForKurtosis(
	listeningPortNumber uint16,
	httpPortNumber uint16,
	sinks logs_aggregator.Sinks,
	shouldEnablePersistentVolumeLogsCollection bool,
) *vectorConfigurationCreator {
	config := newVectorConfig(listeningPortNumber, httpPortNumber, sinks, shouldEnablePersistentVolumeLogsCollection)
	return newVectorConfigurationCreator(config)
}
