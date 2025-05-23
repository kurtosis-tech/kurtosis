package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	rootDirpath = "/fluent-bit"

	////////////////////////--FLUENT BIT CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage        = "fluent/fluent-bit:4.0.0"
	tcpTransportProtocol  = port_spec.TransportProtocol_TCP
	httpTransportProtocol = port_spec.TransportProtocol_TCP

	configDirpathInContainer        = rootDirpath + "/etc"
	configFilepathInContainer       = configDirpathInContainer + "/fluent-bit.conf"
	parserConfigFilepathInContainer = configDirpathInContainer + "/kurtosis-parsers.conf" // create an additional parsers file for ones defined by users in kurtosis config

	//these two values are used for configuring the filesystem buffer. See more here: https://docs.fluentbit.io/manual/administration/buffering-and-storage#filesystem-buffering-to-the-rescue
	filesystemBufferStorageDirpath = configDirpathInContainer + "/storage/"
	inputFilesystemStorageType     = "filesystem"

	configFileTemplateName = "fluentbitConfigFileTemplate"
	configFileTemplate     = `
[SERVICE]
	log_level {{.Service.LogLevel}}
	http_server {{.Service.HttpServerEnabled}}
	http_listen {{.Service.HttpServerHost}}
	http_port {{.Service.HttpServerPort}}
	storage.path {{.Service.StoragePath}}
	parsers_file /fluent-bit/etc/parsers.conf
	parsers_file {{.Service.KurtosisParsersConfigFilepath}}
[INPUT]
	name {{.Input.Name}}
	listen {{.Input.Listen}}
	port {{.Input.Port}}
	storage.type {{.Input.StorageType}}
{{- range .Filters}}
[FILTER]
	name {{.Name}}
	match {{.Match}}
{{- range .Params}}
	{{.Key}} {{.Value}}
{{- end}}{{end}}
[OUTPUT]
	name {{.Output.Name}}
	match {{.Output.Match}}
	host {{.Output.Host}}
	port {{.Output.Port}}
`

	parserConfigFileTemplateName = "fluentbitParserConfigFileTemplate"
	parserConfigFileTemplate     = `{{- range .Parsers}}[PARSER]
{{- range $key, $value := . }}
	{{$key}} {{$value}}
{{- end }}{{ end }}
`

	healthCheckEndpointPath = "api/v1/health"
	////////////////////////--FINISH FLUENT BIT CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--FLUENTBIT CONFIGURATION SECTION--/////////////////////////////
	logLevel               = "debug"
	httpServerEnabledValue = "On"
	httpServerLocalhost    = "0.0.0.0"
	inputName              = "forward"
	inputListenIP          = "0.0.0.0"
	matchAllRegex          = "*"

	// fluentbit doesn't have a dedicated vector output plugin but vector added a source input plugin for fluentbit
	// with the ability to pick up logs over fluentbit's forward output plugin, PR here: https://github.com/vectordotdev/vector/pull/7548
	vectorOutputTypeName = "forward"
	////////////////////////--FINISH FLUENTBIT CONFIGURATION SECTION--/////////////////////////////
)
