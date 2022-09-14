package fluentbit

import (
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"strings"
)

const (
	filterRulesSeparator  = "\n	"
	outputLabelsSeparator = ", "

	renameModifyFilterRuleAction = "rename"

	labelsVarPrefix = "$"

	notAllowedCharInLabels = " .-"
	noSeparationChar       = ""

	shouldChangeNextCharToUpperCaseInitialValue = false
	shouldChangeCharToUpperCaseInitialValue     = false
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
			Name:   inputName,
			Listen: inputListenIP,
			Port:   tcpPortNumber,
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
			TenantIDKey: getTenantIdKeyFromKurtosisLabels(),
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

	kurtosisLabelsForLogs := getTrackedKurtosisLabelsForLogs()
	for _, kurtosisLabel := range kurtosisLabelsForLogs {
		validFormatLabelValue := newValidFormatLabelValue(kurtosisLabel)
		modifyFilterRule := fmt.Sprintf("%v %v %v", renameModifyFilterRuleAction, kurtosisLabel, validFormatLabelValue)
		modifyFilterRules = append(modifyFilterRules, modifyFilterRule)
	}
	return modifyFilterRules
}

func getOutputKurtosisLabelsForLogs() []string {
	outputLabels := []string{}

	kurtosisLabelsForLogs := getTrackedKurtosisLabelsForLogs()
	for _, kurtosisLabel := range kurtosisLabelsForLogs {
		validFormatLabelValue := newValidFormatLabelValue(kurtosisLabel)
		outputLabel := fmt.Sprintf("%v%v", labelsVarPrefix, validFormatLabelValue)
		outputLabels = append(outputLabels, outputLabel)
	}
	return outputLabels
}

func getTrackedKurtosisLabelsForLogs() []string {
	kurtosisLabelsForLogs := []string{
		label_key_consts.GUIDDockerLabelKey.GetString(),
		label_key_consts.ContainerTypeDockerLabelKey.GetString(),
	}
	return kurtosisLabelsForLogs
}

func getTenantIdKeyFromKurtosisLabels() string {
	return label_key_consts.EnclaveIDDockerLabelKey.GetString()
}

func newValidFormatLabelValue(stringToModify string) string {
	stringToModifyInLowerCase := strings.ToLower(stringToModify)
	shouldChangeNextCharToUpperCase := shouldChangeNextCharToUpperCaseInitialValue
	shouldChangeCharToUpperCase := shouldChangeCharToUpperCaseInitialValue
	var newString string
	for _, currenChar := range strings.Split(stringToModifyInLowerCase, noSeparationChar) {
		newChar := currenChar
		shouldChangeCharToUpperCase = shouldChangeNextCharToUpperCase
		if shouldChangeCharToUpperCase {
			newChar = strings.ToUpper(newChar)
		}
		if strings.ContainsAny(currenChar, notAllowedCharInLabels) {
			shouldChangeNextCharToUpperCase = true
		} else {
			shouldChangeNextCharToUpperCase = false
			newString = newString + newChar
		}
	}
	return newString
}
