package fluentbit

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"

const (
	rootDirpath = "/fluent-bit"

	////////////////////////--FLUENTBIT DAEMONSET CONFIGURATION SECTION--/////////////////////////////
	daemonSetName         = "fluent-bit"
	containerImage        = "fluent/fluent-bit:1.9.7"
	tcpTransportProtocol  = port_spec.TransportProtocol_TCP
	httpTransportProtocol = port_spec.TransportProtocol_TCP

	configDirpathInContainer = rootDirpath + "/etc/"
	configFileName           = "fluent-bit.conf"
	parserFileName           = "parsers.conf"

	//these two values are used for configuring the filesystem buffer. See more here: https://docs.fluentbit.io/manual/administration/buffering-and-storage#filesystem-buffering-to-the-rescue
	filesystemBufferStorageDirpath = rootDirpath + "storage/"
	inputFilesystemStorageType     = "filesystem"

	configFileTemplateName = "fluentbitConfigFileTemplate"
	configFileTemplate     = `
[SERVICE]
	log_level {{.Service.LogLevel}}
	http_server {{.Service.HttpServerEnabled}}
	http_listen {{.Service.HttpServerHost}}
	http_port {{.Service.HttpServerPort}}
	storage.path {{.Service.StoragePath}}
	parsers_file ` + configDirpathInContainer + parserFileName + `
[INPUT]
	Name              tail
	Tag               kube.*
	Path              /var/log/containers/*.log
	Parser            docker
	DB                /var/log/flb_kube.db
	Mem_Buf_Limit     5MB
	Skip_Long_Lines   On
	Refresh_Interval  10
[FILTER]
	Name                kubernetes 
	Match               kube.*
	Kube_URL            https://kubernetes.default.svc.cluster.local:443
	Merge_Log           Off
	K8S-Logging.Parser  On
[FILTER]
	name parser
	match *
	parser json
	key_name log
	reserve_data On
[OUTPUT]
	name stdout
	match *
[OUTPUT]
	name {{.Output.Name}}
	match {{.Output.Match}}
	host {{.Output.Host}}
	port {{.Output.Port}}
`

	parserFileContent = `
[PARSER]
	Name   json
	Format json
[PARSER]
	Name        docker
	Format      json
	Time_Key    time
	Time_Format %Y-%m-%dT%H:%M:%S.%L
	Time_Keep   On
	# Command      |  Decoder | Field | Optional Action
	# =============|==================|=================
	Decode_Field_As   escaped    log
`
	kubernetesAppLabel = "k8s-app"
	configMapName      = "fluent-bit-config"

	healthCheckEndpointPath = "api/v1/health"
	////////////////////////--FINISH FLUENTBIT CONTAINER CONFIGURATION SECTION--/////////////////////////////

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
