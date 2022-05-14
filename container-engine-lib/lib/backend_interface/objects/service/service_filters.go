package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type ServiceFilters struct {
	IDs map[ServiceID]bool

	// Disjunctive set of user service GUIDs to find user services for
	// If nil or empty, will match all GUIDs
	GUIDs map[ServiceGUID]bool

	// Disjunctive set of statuses that returned user services must conform to
	// If nil or empty, will match all statuses
	Statuses map[UserServiceStatus]bool

	// Disjunctive set of enclave IDs for which to return user services
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool
}
