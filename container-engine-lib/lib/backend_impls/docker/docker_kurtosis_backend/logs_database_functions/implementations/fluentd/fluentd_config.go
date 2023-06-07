package fluentd

import (
	"bytes"
	"github.com/kurtosis-tech/stacktrace"
	"text/template"
)

const configTemplate = `
<source>
  @type forward
  port {{.HttpPortNumber}}
  bind 0.0.0.0
</source>

<match **>
  @type file
  path ` + dirpath + `data.*.log
  append true
</match>
`

type FluentdConfig struct {
	HttpPortNumber uint16
}

func newDefaultFluentdConfigForKurtosisCentralizedLogs(httpPortNumber uint16) *FluentdConfig {
	return &FluentdConfig{
		HttpPortNumber: httpPortNumber,
	}
}

func (fluentdConfig *FluentdConfig) RenderConfig() (string, error) {
	parsedTemplate, err := template.New("config-template").Parse(configTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to parse Fluentd config template")
	}

	configBuffer := new(bytes.Buffer)
	if err = parsedTemplate.Execute(configBuffer, fluentdConfig); err != nil {
		return "", stacktrace.Propagate(err, "Unable to execute Fluentd config template with the provided data")
	}
	return configBuffer.String(), nil
}
