package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type Filter struct {
	Name   string
	Match  string
	Params map[string]string
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

type FluentbitConfig struct {
	Service *Service
	Input   *Input
	Filters []*Filter
	Output  *Output
}

func newFluentbitConfigForKurtosisCentralizedLogs(
	logsAggregatorHost string,
	logsAggregatorPort uint16,
	tcpPortNumber uint16,
	httpPortNumber uint16,
	logsCollectorFilters []logs_collector.Filter,
) (*FluentbitConfig, error) {
	filters := make([]*Filter, len(logsCollectorFilters))
	for idx, logsCollectorFilter := range logsCollectorFilters {
		params := logsCollectorFilter
		logrus.Infof("params: %v", params)

		name, ok := params["name"]
		if !ok {
			return nil, stacktrace.NewError("name key is required for fluentbit filters")
		}
		match, ok := params["match"]
		if !ok {
			return nil, stacktrace.NewError("match key is required for fluentbit filters")
		}

		paramsWithNoNameAndMatch := make(map[string]string)
		for k, v := range params {
			if k != "name" && k != "match" {
				paramsWithNoNameAndMatch[k] = v
			}
		}

		filters[idx] = &Filter{
			Name:   name,
			Match:  match,
			Params: paramsWithNoNameAndMatch,
		}
	}

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
		Filters: filters,
		Output: &Output{
			Name:  vectorOutputTypeName,
			Match: matchAllRegex,
			Host:  logsAggregatorHost,
			Port:  logsAggregatorPort,
		},
	}, nil
}
