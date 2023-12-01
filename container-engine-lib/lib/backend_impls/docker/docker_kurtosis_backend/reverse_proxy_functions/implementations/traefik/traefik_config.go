package traefik

import (
	"bytes"
	"text/template"

	"github.com/kurtosis-tech/stacktrace"
)

type TraefikConfig struct {
	WebAddress     uint16
	TraefikAddress uint16
	NetworkId      string
}

func newDefaultTraefikConfig(httpPort uint16, dashboardPort uint16) *TraefikConfig {
	return &TraefikConfig{
		WebAddress:     httpPort,
		TraefikAddress: dashboardPort,
		NetworkId:      traefikNetworkid,
	}
}

func (cfg *TraefikConfig) getConfigFileContent() (string, error) {
	cfgFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Traefik's config template.")
	}

	templateStrBuffer := &bytes.Buffer{}
	if err := cfgFileTemplate.Execute(templateStrBuffer, cfg); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing Traefik's config file template.")
	}
	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
