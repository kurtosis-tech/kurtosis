package vector

const (
	configDirpath                = "/etc/vector/"
	healthCheckEndpoint          = "health"
	defaultGraphQlApiHttpPortNum = uint16(8686)
	httpProtocolStr              = "http"

	////////////////////////--VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "timberio/vector:0.31.0-debian"

	configFilepath = configDirpath + "vector.toml"
	binaryFilepath = "/usr/bin/vector"
	configFileFlag = "-c"

	logsStorageDirpath = "/var/log/kurtosis/"
	////////////////////////--FINISH VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--VECTOR CONFIGURATION SECTION--/////////////////////////////
	fluentBitSourceId        = "\"fluent_bit\""
	fluentBitSourceType      = "\"fluent\""
	fluentBitSourceIpAddress = "0.0.0.0"

	fileSinkId = "\"file\""
	fileTypeId = "\"file\""

	// We store log files per-enclave, per-service
	// To construct the filepath, we utilize vectors template syntax that allows us to reference fields in log events
	// https://vector.dev/docs/reference/configuration/template-syntax/
	logsFilepath = "\"" + logsStorageDirpath + "{{ timestamp }}/{{ container_name }}-logs.json\""

	configFileTemplateName = "vectorConfigFileTemplate"
	configFileTemplate     = `
[api]
enabled = true
address = "0.0.0.0:8686"

[sources.{{ .Source.Id }}]
type = {{ .Source.Type }}
address = "{{ .Source.Address }}"

[sinks.{{ .Sink.Id }}]
type = {{ .Sink.Type }}
inputs = {{ .Sink.Inputs }}
path = {{ .Sink.Filepath }}
encoding.codec = "json"
`
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
