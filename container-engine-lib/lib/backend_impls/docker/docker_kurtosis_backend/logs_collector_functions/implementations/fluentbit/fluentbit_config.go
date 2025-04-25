package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
)

type Service struct {
	LogLevel                      string
	HttpServerEnabled             string
	HttpServerHost                string
	HttpServerPort                uint16
	StoragePath                   string
	KurtosisParsersConfigFilepath string
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
	Filters []logs_collector.Filter
	Output  *Output
}

type ParserConfig struct {
	Parsers []logs_collector.Parser
}

func newFluentbitConfigForKurtosisCentralizedLogs(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
) (*FluentbitConfig, *ParserConfig) {
	return &FluentbitConfig{
			Service: &Service{
				LogLevel:                      logLevel,
				HttpServerEnabled:             httpServerEnabledValue,
				HttpServerHost:                httpServerLocalhost,
				HttpServerPort:                httpPortNumber,
				StoragePath:                   filesystemBufferStorageDirpath,
				KurtosisParsersConfigFilepath: parserConfigFilepathInContainer,
			},
			Input: &Input{
				Name:        inputName,
				Listen:      inputListenIP,
				Port:        tcpPortNumber,
				StorageType: inputFilesystemStorageType,
			},
			Filters: logsCollectorFilters,
			Output: &Output{
				Name:  vectorOutputTypeName,
				Match: matchAllRegex,
				Host:  logsAggregatorHost,
				Port:  logsAggregatorPort,
			},
		}, &ParserConfig{
			Parsers: logsCollectorParsers,
		}
}
