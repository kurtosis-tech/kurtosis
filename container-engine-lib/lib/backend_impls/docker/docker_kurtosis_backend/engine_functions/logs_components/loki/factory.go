package loki

func CreateLokiConfiguredForKurtosis() *Loki {
	config := newDefaultConfigForKurtosisCentralizedLogs()
	loki := NewLoki(config)
	return loki
}