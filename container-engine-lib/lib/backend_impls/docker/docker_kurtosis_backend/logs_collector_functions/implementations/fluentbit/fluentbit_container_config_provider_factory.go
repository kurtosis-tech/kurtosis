package fluentbit

func createFluentbitContainerConfigProviderForKurtosis(
	logsDatabaseHost string,
	logsDatabasePort uint16,
	httpPortNumber uint16,
) *fluentbitContainerConfigProvider {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(logsDatabaseHost, logsDatabasePort, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitContainerConfigProvider(config, httpPortNumber)
	return fluentbitContainerConfigProvider
}
