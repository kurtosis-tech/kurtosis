package fluentbit

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"text/template"
)

type fluentbitConfigurationCreator struct {
	config *FluentbitConfig
}

func newFluentbitConfigurationCreator(config *FluentbitConfig) *fluentbitConfigurationCreator {
	return &fluentbitConfigurationCreator{config: config}
}

func (fluent *fluentbitConfigurationCreator) CreateConfiguration(
	ctx context.Context,
	engineNamespace string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	fluentbitConfigContent, err := fluent.getConfigFileContent()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to render Fluentbit config content")
	}
	parsersConfigContent := parserFileContent

	_, err = kubernetesManager.CreateConfigMap(
		ctx,
		engineNamespace,
		configMapName,
		map[string]string{
			kubernetesAppLabel: daemonSetName,
		},
		map[string]string{},
		map[string][]byte{
			configFileName: []byte(fluentbitConfigContent),
			parserFileName: []byte(parsersConfigContent),
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create the Kubernetes ConfigMap required to start Fluentbit logs collector")
	}
	return nil
}

func (fluent *fluentbitConfigurationCreator) getConfigFileContent() (string, error) {

	cngFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Fluentbit config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	if err := cngFileTemplate.Execute(templateStrBuffer, fluent.config); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing the Fluentbit config file template")
	}

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
