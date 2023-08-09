package fluentbit

type FluentbitConfig struct {
	Service *Service
	Input   *Input
	Output  *Output
}

type Service struct {
	LogLevel          string
	HttpServerEnabled string
	HttpServerHost    string
	HttpServerPort    uint16
	StoragePath       string
}

type Input struct {
	Name        string
	Listen      string
	Port        uint16
	StorageType string
}

type Output struct {
	Name  string
	Match string
	Host  string
	Port  uint16
}

func newDefaultFluentbitConfigForKurtosisCentralizedLogs(
	logAggregatorHost string,
	logAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
) *FluentbitConfig {
	return &FluentbitConfig{
		Service: &Service{
			LogLevel:          logLevel,
			HttpServerEnabled: httpServerEnabledValue,
			HttpServerHost:    httpServerLocalhost,
			HttpServerPort:    httpPortNumber,
			StoragePath:       filesystemBufferStorageDirpath,
		},
		Input: &Input{
			Name:        inputName,
			Listen:      inputListenIP,
			Port:        tcpPortNumber,
			StorageType: inputFilesystemStorageType,
		},
		Output: &Output{
			Name:  vectorOutputTypeName,
			Match: matchAllRegex,
			Host:  logAggregatorHost,
			Port:  logAggregatorPort,
		},
	}
}
