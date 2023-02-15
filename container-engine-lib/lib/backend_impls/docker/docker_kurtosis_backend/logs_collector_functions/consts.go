package logs_collector_functions

import "time"

//Centralized logs component port IDs

const (
	logsCollectorTcpPortId  = "tcp"
	logsCollectorHttpPortId = "http"

	stopLogsCollectorContainersTimeout = 1 * time.Minute
)
