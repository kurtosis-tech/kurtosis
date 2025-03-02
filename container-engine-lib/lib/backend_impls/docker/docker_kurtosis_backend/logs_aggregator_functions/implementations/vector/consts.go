package vector

const (
	configDirpath = "/etc/vector/"

	////////////////////////--VECTOR CONTAINER CONFIGURATION SECTION--/////////////////////////////
	containerImage = "timberio/vector:0.45.0-debian"

	configFilepath = configDirpath + "vector.json"
	binaryFilepath = "/usr/bin/vector"
	configFileFlag = "-c"

	logsStorageDirpath = "/var/log/kurtosis/"
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
