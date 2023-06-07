package fluentd

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	////////////////////////--FLUENTD CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage        = "fluent/fluentd:stable"
	httpTransportProtocol = port_spec.TransportProtocol_TCP
	logLevel              = "info"

	configDirpath  = "/fluentd/etc/"
	configFilepath = configDirpath + "fluentd.conf"
	binaryFilepath = "fluentd"
	configFileFlag = "-c"
	////////////////////////--FINISH FLUENTD CONTAINER CONFIGURATION SECTION--/////////////////////////////

	dirpath = "/logs/"
)
