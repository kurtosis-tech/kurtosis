package fluentbit

func CreateFluentbitConfiguredForKurtosis() *Fluentbit {
	config := newDefaultConfigForKurtosisCentralizedLogs()
	fluentbit := NewFluentbit(config)
	return fluentbit
}
