package fluentbit

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

func createFluentbitContainerConfigProviderForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
) *fluentbitContainerConfigProvider {
	configCreator := createFluentbitConfigurationCreatorForKurtosis(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber, logsCollectorFilters, logsCollectorParsers)
	return newFluentbitContainerConfigProvider(configCreator.config, tcpPortNumber, httpPortNumber)
}
