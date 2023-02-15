package loki

func createLokiContainerConfigProviderForKurtosis(httpPortNumber uint16) *lokiContainerConfigProvider {
	lokiConfig := newDefaultLokiConfigForKurtosisCentralizedLogs(httpPortNumber)
	lokiContainerConfigProviderObj := newLokiContainerConfigProvider(lokiConfig, httpPortNumber)
	return lokiContainerConfigProviderObj
}
