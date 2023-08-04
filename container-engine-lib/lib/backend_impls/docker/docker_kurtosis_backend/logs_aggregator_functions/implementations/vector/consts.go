package vector

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	configDirpath = "/etc/vector/"

	////////////////////////--VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage        = "timberio/vector:0.31.0-debian"
	httpTransportProtocol = port_spec.TransportProtocol_TCP
	logLevel              = "info"

	configFilepath = configDirpath + "vector.toml"
	binaryFilepath = "/usr/bin/vector"
	configFileFlag = "--config"
	////////////////////////--FINISH VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--VECTOR CONFIGURATION SECTION--/////////////////////////////
	fluentBitSourceId        = "fluent_bit"
	fluentBitSourceType      = "fluent"
	fluentBitSourceIpAddress = "0.0.0.0"
	fluentBitSourcePort      = "9000"

	stdoutSinkID = "stdout"
	stdoutTypeId = "console"
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
