package vector

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"strconv"
	"text/template"
)

type VectorConfig struct {
	Source *Source
	Sinks  []*Sink
}

type Source struct {
	Id      string
	Type    string
	Address string
}

type Sink struct {
	Id       string
	Type     string
	Inputs   []string
	Filepath string
}

func newDefaultVectorConfig(listeningPortNumber uint16) *VectorConfig {
	return &VectorConfig{
		Source: &Source{
			Id:      fluentBitSourceId,
			Type:    fluentBitSourceType,
			Address: fmt.Sprintf("%s:%s", fluentBitSourceIpAddress, strconv.Itoa(int(listeningPortNumber))),
		},
		Sinks: []*Sink{
			{
				Id:       "uuid_" + fileSinkIdSuffix,
				Type:     fileTypeId,
				Inputs:   []string{fluentBitSourceId},
				Filepath: VectorLogsFilepathFormat,
			},
		},
	}
}

func (cfg *VectorConfig) getConfigFileContent() (string, error) {
	srcCfgFileTemplate, err := template.New(sourceConfigFileTemplateName).Parse(srcConfigFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Vector's source config template.")
	}
	sinkCfgFileTemplate, err := template.New(sinkConfigFileTemplateName).Parse(sinkConfigFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Vector's sink config template.")
	}

	templateStrBuffer := &bytes.Buffer{}

	if err := srcCfgFileTemplate.Execute(templateStrBuffer, cfg.Source); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing Vector's source config file template.")
	}
	for _, sink := range cfg.Sinks {
		if err := sinkCfgFileTemplate.Execute(templateStrBuffer, sink); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred executing Vector's sink config file template.")
		}
	}

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
