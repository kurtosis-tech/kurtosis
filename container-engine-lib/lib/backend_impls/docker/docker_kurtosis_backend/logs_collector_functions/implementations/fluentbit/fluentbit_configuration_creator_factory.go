package fluentbit

func createFluentbitConfigurationCreatorForKurtosis(
	logsDatabaseHost string,
	logsDatabasePort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *fluentbitConfigurationCreator {
	config := newDefaultFluentbitConfigForKurtosisCentralizedLogs(logsDatabaseHost, logsDatabasePort, tcpPortNumber, httpPortNumber)
	fluentbitContainerConfigProvider := newFluentbitConfigurationCreator(config)
	return fluentbitContainerConfigProvider
}

