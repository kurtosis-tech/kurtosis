package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

func createFluentbitContainerConfigProviderForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
) *fluentbitContainerConfigProvider {
	configCreator := createFluentbitConfigurationCreatorForKurtosis(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber, logsCollectorFilters)
	return newFluentbitContainerConfigProvider(configCreator.config, tcpPortNumber, httpPortNumber)
}
