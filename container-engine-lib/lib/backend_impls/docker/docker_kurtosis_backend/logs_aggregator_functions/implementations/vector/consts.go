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

	fileSinkIdSuffix = "file"
	fileTypeId       = "\"file\""

	// To construct the filepath, we utilize vectors template syntax that allows us to reference fields in log events
	// https://vector.dev/docs/reference/configuration/template-syntax/
	baseLogsFilepath = "\"" + logsStorageDirpath + "%%Y/%%V/%%u/%%H/"

	uuidLogsFilepath = baseLogsFilepath + "{{ enclave_uuid }}/{{ service_uuid }}.json\""

	sourceConfigFileTemplateName = "srcVectorConfigFileTemplate"
	sinkConfigFileTemplateName   = "sinkVectorConfigFileTemplate"

	// Note: we set buffer to block so that we don't drop any logs, however this could apply backpressure up the topology
	// if we start noticing slowdown due to vector buffer blocking, we might want to revisit our architecture
	srcConfigFileTemplate = `
[sources.{{ .Id }}]
type = {{ .Type }}
address = "{{ .Address }}"
`
	sinkConfigFileTemplate = `
[sinks.{{ .Id }}]
type = {{ .Type }}
inputs = {{ .Inputs }}
path = {{ .Filepath }}	
encoding.codec = "json"
buffer.when_full = "block"
`
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
