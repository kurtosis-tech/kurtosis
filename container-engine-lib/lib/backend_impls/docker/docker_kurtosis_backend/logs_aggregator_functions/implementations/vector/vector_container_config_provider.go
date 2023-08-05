package vector

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"text/template"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
)

type vectorContainerConfigProvider struct {
	config *VectorConfig
}

func newVectorContainerConfigProvider(config *VectorConfig) *vectorContainerConfigProvider {
	return &vectorContainerConfigProvider{config: config}
}

func (vector *vectorContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	logsAggregatorVolumeName string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	volumeMounts := map[string]string{
		logsAggregatorVolumeName: configDirpath,
	}

	logsAggregatorConfigContentStr, err := vector.getConfigFileContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	// Create cmd to
	// 1. create config file in appropriate location in logs aggregator container
	// 2. start the logs aggregator with the config file
	overrideCmd := []string{
		fmt.Sprintf(
			"%v '%v' > %v && %v %v %v",
			printfCmdName,
			logsAggregatorConfigContentStr,
			configFilepath,
			binaryFilepath,
			configFileFlag,
			configFilepath,
		),
	}
	logrus.Debugf("OVERRIDE CMD LOG AGGREGATOR: %v", overrideCmd)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithVolumeMounts(
		volumeMounts,
	).WithEntrypointArgs(
		[]string{
			shBinaryFilepath,
		},
	).WithCmdArgs(
		overrideCmd,
	).Build()

	return createAndStartArgs, nil
}

func (vector *vectorContainerConfigProvider) getConfigFileContent() (string, error) {
	cngFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Vector config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	if err := cngFileTemplate.Execute(templateStrBuffer, vector.config); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing the Vector config file template")
	}

	templateStr := templateStrBuffer.String()
	logrus.Debugf("VECTOR CONFIG FILE: %s", templateStr)

	return templateStr, nil
}
