package logs_database

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"

type LogsDatabaseFilters struct {
	// The status that returned logs database must conform to
	// If nil or empty, will match all statuses
	Status container_status.ContainerStatus
}
