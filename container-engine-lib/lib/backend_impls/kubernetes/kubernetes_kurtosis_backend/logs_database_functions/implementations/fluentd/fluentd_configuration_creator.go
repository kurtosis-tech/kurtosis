package fluentd

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/stacktrace"
	"text/template"
)

type fluentdConfigurationCreator struct {
	config *FluentdConfig
}

func newFluentdConfigurationCreator(config *FluentdConfig) *fluentdConfigurationCreator {
	return &fluentdConfigurationCreator{config: config}
}

func (fluent *fluentdConfigurationCreator) CreateConfiguration(
	ctx context.Context,
	engineNamespace string,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) error {
	fluentbitConfigContent, err := fluent.getConfigFileContent()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to render Fluentd config content")
	}

	_, err = kubernetesManager.CreateConfigMap(
		ctx,
		engineNamespace,
		configMapName,
		map[string]string{
			kubernetesAppLabel: fluentdName,
		},
		map[string]string{},
		map[string][]byte{
			configFileName: []byte(fluentbitConfigContent),
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create the Kubernetes ConfigMap required to start Fluentbit logs collector")
	}
	return nil
}

func (fluent *fluentdConfigurationCreator) getConfigFileContent() (string, error) {

	cngFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Fluentd config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	if err := cngFileTemplate.Execute(templateStrBuffer, fluent.config); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing the Fluentd config file template")
	}

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
