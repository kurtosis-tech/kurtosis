package fluentbit

func CreateFluentbitContainerConfigProviderForKurtosis(
	logsDatabaseHost string,
	logsDatabasePort uint16,
	httpPortNumber uint16,
) *FluentbitContainerConfigProvider {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(logsDatabaseHost, logsDatabasePort, httpPortNumber)
	fluentbitContainerConfigProvider := NewFluentbitContainerConfigProvider(config, httpPortNumber)
	return fluentbitContainerConfigProvider
}
