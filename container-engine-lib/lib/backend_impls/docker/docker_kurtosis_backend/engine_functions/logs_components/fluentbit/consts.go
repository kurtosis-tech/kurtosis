package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	rootDirpath = "/fluent-bit"

	////////////////////////--LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage          = "fluent/fluent-bit:1.9.7"
	tcpPortNumber    uint16 = 24224 // Default Fluentbit TCP port number, more here: https://docs.fluentbit.io/manual/pipeline/outputs/forward
	tcpPortProtocol         = port_spec.PortProtocol_TCP
	httpPortProtocol        = port_spec.PortProtocol_TCP

	lokiOutputTypeName = "loki"

	configDirpathInContainer  = rootDirpath + "/etc"
	configFilepathInContainer = configDirpathInContainer + "/fluent-bit.conf"

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
[INPUT]
	name {{.Input.Name}}
	listen {{.Input.Listen}}
	port {{.Input.Port}}
	storage.type  {{.Input.StorageType}}
[FILTER]
	name {{.Filter.Name}}
	match {{.Filter.Match}}
	{{.Filter.GetRulesStr}}
[OUTPUT]
	name {{.Output.Name}}
	match {{.Output.Match}}
	host {{.Output.Host}}
	port {{.Output.Port}}
	labels {{.Output.GetLabelsStr}}
	line_format {{.Output.LineFormat}}
	tenant_id_key {{.Output.TenantIDKey}}
	retry_limit {{.Output.RetryLimit}}
`

	healthCheckEndpointPath = "api/v1/health"
	////////////////////////--FINISH LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--FLUENTBIT CONFIGURATION SECTION--/////////////////////////////
	logLevel               = "debug"
	httpServerEnabledValue = "On"
	httpServerLocalhost    = "0.0.0.0"
	inputName              = "forward"
	inputListenIP          = "0.0.0.0"
	modifyFilterName       = "modify"
	matchAllRegex          = "*"
	jsonLineFormat         = "json"
	unlimitedOutputRetry   = "no_limits"
	////////////////////////--FINISH FLUENTBIT CONFIGURATION SECTION--/////////////////////////////
)
