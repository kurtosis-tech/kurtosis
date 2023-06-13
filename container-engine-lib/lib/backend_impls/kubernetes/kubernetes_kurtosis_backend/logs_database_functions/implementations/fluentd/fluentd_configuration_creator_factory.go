package fluentd

func createFluentdConfigurationCreatorForKurtosis(
	portNumber uint16,
) *fluentdConfigurationCreator {
	config := newDefaultFluentdConfigForKurtosisCentralizedLogs(portNumber)
	fluentdContainerConfigCreator := newFluentdConfigurationCreator(config)
	return fluentdContainerConfigCreator
}
