package fluentbit

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"

const (

	rootDirpath = "/fluent-bit"

	////////////////////////--LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage          = "fluent/fluent-bit:1.9.7-debug"
	tcpPortNumber   uint16 = 24224 // Default Fluentbit TCP port number, more here: https://docs.fluentbit.io/manual/pipeline/outputs/forward
	tcpPortProtocol        = port_spec.PortProtocol_TCP

	binaryFilepath = rootDirpath + "/bin/fluent-bit"
	configFilepath = rootDirpath + "/etc/fluent-bit.conf"
	configFileFlag = "--config"

	configContentTemplate = `
		[INPUT]
			name        forward
			listen      0.0.0.0
			port        24224
		[FILTER]
			name modify
			match *
			rename com.kurtosistech.guid kurtosisGUID
			rename com.kurtosistech.container-type kurtosisContainerType`
	/*configContentTemplate = `
		[INPUT]
			name        forward
			listen      0.0.0.0
			port        24224
		[FILTER]
			name modify
			match *
			rename com.kurtosistech.guid kurtosisGUID
			rename com.kurtosistech.container-type kurtosisContainerType
		[OUTPUT]
			name loki
			match *
			host ${LOKI_HOST}
			port ${LOKI_PORT}
			remove_keys source
			labels job=fluent-bit, $kurtosisContainerType, $kurtosisGUID
			line_format json
			tenant_id_key com.kurtosistech.enclave-id`*/
	////////////////////////--FINISH LOKI CONTAINER CONFIGURATION SECTION--/////////////////////////////
)
