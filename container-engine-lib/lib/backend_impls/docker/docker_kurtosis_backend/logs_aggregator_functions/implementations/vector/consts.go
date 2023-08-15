package vector

const (
	configDirpath = "/etc/vector/"

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
	//logsFilepath    = "\"" + logsStorageDirpath + "/{{ container_name }}.json\""
	logsFilepath = "\"" + logsStorageDirpath + "logs.json\""

	configFileTemplateName = "vectorConfigFileTemplate"
	configFileTemplate     = `
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
