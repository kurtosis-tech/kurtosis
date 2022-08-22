package loki

func CreateLokiContainerConfigProviderForKurtosis() *LokiContainerConfigProvider {
	lokiConfig := newDefaultLokiConfigForKurtosisCentralizedLogs()
	lokiContainerConfigProvider := NewLokiContainerConfigProvider(lokiConfig)
	return lokiContainerConfigProvider
}