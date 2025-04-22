package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

func createFluentbitConfigurationCreatorForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
) *fluentbitConfigurationCreator {
	config := newFluentbitConfigForKurtosisCentralizedLogs(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber, logsCollectorFilters, logsCollectorParsers)
	fluentbitContainerConfigCreator := newFluentbitConfigurationCreator(config)
	return fluentbitContainerConfigCreator
}
