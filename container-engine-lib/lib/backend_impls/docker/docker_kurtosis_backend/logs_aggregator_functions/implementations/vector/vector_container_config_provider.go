package vector

import (
	"bytes"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"text/template"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
	shCmdFlag        = "-c"
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
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {
	logsAggregatorConfigContentStr, err := vector.getConfigFileContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	// Create cmd to
	// 1. create config file in appropriate location in logs aggregator container
	// 2. start the logs aggregator with the config file
	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v '%v' > %v && %v %v=%v",
			printfCmdName,
			logsAggregatorConfigContentStr,
			configFilepath,
			binaryFilepath,
			configFileFlag,
			configFilepath,
		),
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
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

	return templateStr, nil
}
