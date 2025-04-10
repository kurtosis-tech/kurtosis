package vector

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
)

func createVectorConfigurationCreatorForKurtosis(
	listeningPortNumber uint16,
	httpPortNumber uint16,
	sinks logs_aggregator.Sinks,
	shouldTurnOffPersistentVolumeLogsCollection bool,
) *vectorConfigurationCreator {
	config := newVectorConfig(listeningPortNumber, httpPortNumber, sinks, shouldTurnOffPersistentVolumeLogsCollection)
	return newVectorConfigurationCreator(config)
}
