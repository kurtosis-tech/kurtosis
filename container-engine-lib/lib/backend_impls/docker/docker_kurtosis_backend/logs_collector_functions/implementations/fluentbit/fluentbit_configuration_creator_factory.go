package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
)

func createFluentbitConfigurationCreatorForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
) (*fluentbitConfigurationCreator, error) {
	config, err := newFluentbitConfigForKurtosisCentralizedLogs(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber, logsCollectorFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Fluentbit config")
	}
	fluentbitContainerConfigCreator := newFluentbitConfigurationCreator(config)
	return fluentbitContainerConfigCreator, nil
}
