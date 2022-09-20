package service

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
)

// Selector for matching services inside an enclave
type ServiceFilters struct {
	IDs map[ServiceID]bool

	// Disjunctive set of user service GUIDs to find user services for
	// If nil or empty, will match all GUIDs in the enclave
	GUIDs map[ServiceGUID]bool

	// Disjunctive set of statuses that returned user services must conform to
	// If nil or empty, will match all statuses
	Statuses map[container_status.ContainerStatus]bool
}