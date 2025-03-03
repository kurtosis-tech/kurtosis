package vector

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
)

const (
	configDirpath = "/etc/vector/"

	////////////////////////--VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "timberio/vector:0.45.0-debian"

	configFilepath = configDirpath + "vector.yaml"
	binaryFilepath = "/usr/bin/vector"
	configFileFlag = "-c"

	logsStorageDirpath      = "/var/log/kurtosis/"
	healthCheckEndpointPath = "health"
	httpTransportProtocol   = port_spec.TransportProtocol_TCP
	////////////////////////--FINISH VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////

	////////////////////////--VECTOR CONFIGURATION SECTION--/////////////////////////////
	defaultSourceId          = "kurtosis_default_source"
	fluentBitSourceType      = "fluent"
	fluentBitSourceIpAddress = "0.0.0.0"
	fileSinkType             = "file"

	// We instruct vector to store log files per-year, per-week (00-53), per-enclave, per-service
	// To construct the filepath, we utilize vectors template syntax that allows us to reference fields in log events
	// https://vector.dev/docs/reference/configuration/template-syntax/
	baseLogsFilepath = "\"" + logsStorageDirpath + "%%G/%%V/"

	uuidLogsFilepath = baseLogsFilepath + "{{ enclave_uuid }}/{{ service_uuid }}.json\""
	////////////////////////--FINISH--VECTOR CONFIGURATION SECTION--/////////////////////////////
)
