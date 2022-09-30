package logs_collector

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"

type LogsCollectorFilters struct {
	// The status that returned logs collector must conform to
	// If nil or empty, will match all statuses
	Status container_status.ContainerStatus
}
