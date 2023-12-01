package reverse_proxy

import (
	"bytes"
	"text/template"

	"github.com/kurtosis-tech/stacktrace"
)

type ReverseProxyConfig struct {
	HttpPort      uint16
	DashboardPort uint16
	NetworkId     string
}

func NewDefaultReverseProxyConfig(httpPort uint16, dashboardPort uint16, networkId string) *ReverseProxyConfig {
	return &ReverseProxyConfig{
		HttpPort:      httpPort,
		DashboardPort: dashboardPort,
		NetworkId:     networkId,
	}
}

func (cfg *ReverseProxyConfig) GetConfigFileContent(configFileTemplate string) (string, error) {
	cfgFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the reverse proxy's config template.")
	}

	templateStrBuffer := &bytes.Buffer{}
	if err := cfgFileTemplate.Execute(templateStrBuffer, cfg); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing the reverse proxy's config file template.")
	}
	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
