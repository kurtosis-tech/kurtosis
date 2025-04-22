package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

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

type FluentbitConfig struct {
	Service *Service
	Input   *Input
	Parsers []logs_collector.Parser
	Filters []logs_collector.Filter
	Output  *Output
}

func newFluentbitConfigForKurtosisCentralizedLogs(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
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
		Parsers: logsCollectorParsers,
		Filters: logsCollectorFilters,
		Output: &Output{
			Name:  vectorOutputTypeName,
			Match: matchAllRegex,
			Host:  logsAggregatorHost,
			Port:  logsAggregatorPort,
		},
	}
}
