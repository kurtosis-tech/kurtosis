package fluentbit

func CreateFluentbitConfiguredForKurtosis(
	logsDatabaseHost string,
	logsDatabasePort uint16,
) *Fluentbit {
	config := newDefaultConfigForKurtosisCentralizedLogsForDocker(logsDatabaseHost, logsDatabasePort)
	fluentbit := NewFluentbit(config)
	return fluentbit
}
