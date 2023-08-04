package fluentbit

func createFluentbitContainerConfigProviderForKurtosis(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *fluentbitContainerConfigProvider {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(logsAggregatorHost, logsAggregatorPort, tcpPortNumber, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitContainerConfigProvider(config, tcpPortNumber, httpPortNumber)
	return fluentbitContainerConfigProvider
}
