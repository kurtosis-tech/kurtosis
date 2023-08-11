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

	logsStorageDirpath = "/tmp/"
	////////////////////////--FINISH VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--VECTOR CONFIGURATION SECTION--/////////////////////////////
	fluentBitSourceId        = "\"fluent_bit\""
	fluentBitSourceType      = "\"fluent\""
	fluentBitSourceIpAddress = "0.0.0.0"

	// TODO: change output when persistent volume is implemented
	stdoutSinkID = "\"stdout\""
	stdoutTypeId = "\"console\""

	fileSinkId      = "\"file\""
	fileTypeId      = "\"file\""
	filepathForLogs = "/tmp/vector.txt"

	configFileTemplateName = "vectorConfigFileTemplate"
	//	configFileTemplate     = `
	//[api]
	//enabled = true
	//address = "0.0.0.0:8686"
	//
	//[sources.{{ .Source.Id }}]
	//type = {{ .Source.Type }}
	//address = "{{ .Source.Address }}"
	//
	//[sinks.{{ .Sink.Id }}]
	//type = {{ .Sink.Type }}
	//inputs = {{ .Sink.Inputs }}
	//encoding.codec = "json"

	configFileTemplate = `
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
`
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
