package loki

func createLokiContainerConfigProviderForKurtosis() *lokiContainerConfigProvider {
	lokiConfig := newDefaultLokiConfigForKurtosisCentralizedLogs()
	lokiContainerConfigProviderObj := newLokiContainerConfigProvider(lokiConfig)
	return lokiContainerConfigProviderObj
}