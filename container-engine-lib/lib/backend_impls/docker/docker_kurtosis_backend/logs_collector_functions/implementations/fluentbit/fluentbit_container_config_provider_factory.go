package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func createFluentbitContainerConfigProviderForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
) (*fluentbitContainerConfigProvider, error) {
	configCreator, err := createFluentbitConfigurationCreatorForKurtosis(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber, logsCollectorFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Fluentbit configuration creator")
	}
	return newFluentbitContainerConfigProvider(configCreator.config, tcpPortNumber, httpPortNumber), nil
}
