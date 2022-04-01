package repl

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
)

type ReplFilters struct {
	// Disjunctive set of enclave IDs for which to return repls
	// If nil or empty, will match all enclave IDs
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive set of repl IDs to find repls for
	// If nil or empty, will match all IDs
	GUIDs map[ReplGUID]bool

	// Disjunctive set of statuses that returned repls must conform to
	// If nil or empty, will match all IDs
	Statuses map[container_status.ContainerStatus]bool
}

