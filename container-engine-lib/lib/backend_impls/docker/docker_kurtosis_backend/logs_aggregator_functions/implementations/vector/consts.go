package vector

const (
	configDirpath = "/etc/vector/"

	////////////////////////--VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "timberio/vector:0.31.0-debian"

	configFilepath = configDirpath + "vector.toml"
	binaryFilepath = "/usr/bin/vector"
	configFileFlag = "--config"
	////////////////////////--FINISH VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--VECTOR CONFIGURATION SECTION--/////////////////////////////
	fluentBitSourceId        = "fluent_bit"
	fluentBitSourceType      = "fluent"
	fluentBitSourceIpAddress = "0.0.0.0"

	stdoutSinkID = "stdout"
	stdoutTypeId = "console"

	configFileTemplateName = "vectorConfigFileTemplate"
	configFileTemplate     = `
[sources.{{ .Source.Id }}]
type = {{ .Source.Type }}
address = {{ .Source.Address }}

[sinks.{{ .Sink.Id }}]
type = {{ .Sink.Type }}
inputs = {{ .Sink.Inputs }}
encoding.codec = "json"
`
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
