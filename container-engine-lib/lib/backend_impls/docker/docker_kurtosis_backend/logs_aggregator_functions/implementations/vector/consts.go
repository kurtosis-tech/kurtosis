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

	// We store log files in the volume per-enclave, per-service
	// To construct the filepath, we utilize vectors template syntax that allows us to reference fields in log events
	// https://vector.dev/docs/reference/configuration/template-syntax/
	logsFilepath = "\"" + logsStorageDirpath + "{{ enclave_uuid }}/{{ service_uuid }}.json\""

	configFileTemplateName = "vectorConfigFileTemplate"

	// Note: we set buffer to block so that we don't drop any logs, however this could apply backpressure up the topology
	// if we start noticing slowdown due to vector buffer blocking, we might want to revisit our architecture
	configFileTemplate = `
[sources.{{ .Source.Id }}]
type = {{ .Source.Type }}
address = "{{ .Source.Address }}"

[sinks.{{ .Sink.Id }}]
type = {{ .Sink.Type }}
inputs = {{ .Sink.Inputs }}
path = {{ .Sink.Filepath }}
encoding.codec = "json"
buffer.when_full = "block"
`

	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
