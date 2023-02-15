package service

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
)

// Selector for matching services inside an enclave
type ServiceFilters struct {
	Names map[ServiceName]bool

	// Disjunctive set of user service UUIDs to find user services for
	// If nil or empty, will match all UUIDs in the enclave
	UUIDs map[ServiceUUID]bool

	// Disjunctive set of statuses that returned user services must conform to
	// If nil or empty, will match all statuses
	Statuses map[container_status.ContainerStatus]bool
}
