package loki

func CreateLokiConfiguredForKurtosis() *Loki {
	lokiConfig := newDefaultLokiConfigForKurtosisCentralizedLogs()
	loki := NewLoki(lokiConfig)
	return loki
}