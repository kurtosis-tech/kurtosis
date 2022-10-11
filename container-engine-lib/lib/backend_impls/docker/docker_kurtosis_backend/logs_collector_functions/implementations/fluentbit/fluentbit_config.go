package fluentbit

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/logs_database_functions/implementations/loki/tags"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_labels_for_logs"
	"strings"
)

const (
	filterRulesSeparator  = "\n	"
	outputLabelsSeparator = ", "

	renameModifyFilterRuleAction = "rename"

	//This is the "record accesor" character used by Fluentbit to dinamically get content from
	//a log stream in JSON format
	labelsVarPrefix = "$"


)

type FluentbitConfig struct {
	Service *Service
	Input   *Input
	Filter  *Filter
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
	Name   string
	Listen string
	Port   uint16
	StorageType string
}

type Filter struct {
	Name  string
	Match string
	Rules []string
}

type Output struct {
	Name        string
	Match       string
	Host        string
	Port        uint16
	Labels      []string
	LineFormat  string
	TenantIDKey string
	RetryLimit  string
}

func newDefaultFluentbitConfigForKurtosisCentralizedLogs(
	lokiHost string,
	lokiPort uint16,
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
		Filter: &Filter{
			Name:  modifyFilterName,
			Match: matchAllRegex,
			Rules: getModifyFilterRulesKurtosisLabels(),
		},
		Output: &Output{
			Name:        lokiOutputTypeName,
			Match:       matchAllRegex,
			Host:        lokiHost,
			Port:        lokiPort,
			Labels:      getOutputKurtosisLabelsForLogs(),
			LineFormat:  jsonLineFormat,
			TenantIDKey: docker_labels_for_logs.LogsDatabaseKurtosisTrackedDockerLabelUsedForIdentifyTenants.GetString(),
			RetryLimit:  unlimitedOutputRetry,
		},
	}
}

func (filter *Filter) GetRulesStr() string {
	return strings.Join(filter.Rules, filterRulesSeparator)
}

func (output *Output) GetLabelsStr() string {
	return strings.Join(output.Labels, outputLabelsSeparator)
}

func getModifyFilterRulesKurtosisLabels() []string {

	modifyFilterRules := []string{}

	kurtosisValidLokiTagsByDockerLabelKey := docker_kurtosis_backend.GetAllLogsDatabaseKurtosisTrackedValidLokiTagsByDockerLabelKey()
	for kurtosisDockerLabelKey, kurtosisValidLokiTag := range kurtosisValidLokiTagsByDockerLabelKey {
		kurtosisDockerLabelKeyStr := kurtosisDockerLabelKey.GetString()
		modifyFilterRule := fmt.Sprintf("%v %v %v", renameModifyFilterRuleAction, kurtosisDockerLabelKeyStr, kurtosisValidLokiTag)
		modifyFilterRules = append(modifyFilterRules, modifyFilterRule)
	}
	return modifyFilterRules
}

func getOutputKurtosisLabelsForLogs() []string {
	outputLabels := []string{}

	kurtosisLokiTagsForLogs := docker_kurtosis_backend.GetAllLogsDatabaseKurtosisTrackedValidLokiTags()
	for _, kurtosisLokiTag := range kurtosisLokiTagsForLogs {
		outputLabel := fmt.Sprintf("%v%v", labelsVarPrefix, kurtosisLokiTag)
		outputLabels = append(outputLabels, outputLabel)
	}
	return outputLabels
}
