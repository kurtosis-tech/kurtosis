package fluentd

type FluentdConfig struct {
	portNumber uint16
}

func newDefaultFluentdConfigForKurtosisCentralizedLogs(portNumber uint16) *FluentdConfig {
	return &FluentdConfig{
		portNumber: portNumber,
	}
}
