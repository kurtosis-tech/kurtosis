package fluentbit

func createFluentbitContainerConfigProviderForKurtosis(
	logsDatabaseHost string,
	logsDatabasePort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *fluentbitContainerConfigProvider {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(logsDatabaseHost, logsDatabasePort, tcpPortNumber, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitContainerConfigProvider(config, tcpPortNumber, httpPortNumber)
	return fluentbitContainerConfigProvider
}
