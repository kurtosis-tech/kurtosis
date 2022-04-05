package module

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type ModuleFilters struct {
	// Disjunctive set of enclave IDs for which to return modules
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive set of module GUIDs to find modules for
	// If nil or empty, will match all GUIDs
	GUIDs map[ModuleGUID]bool

	// Disjunctive set of statuses that returned API containers must conform to
	// If nil or empty, will match all IDs
	Statuses map[container_status.ContainerStatus]bool
}
