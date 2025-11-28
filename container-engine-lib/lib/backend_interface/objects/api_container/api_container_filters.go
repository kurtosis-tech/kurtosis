package api_container

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
)

type APIContainerFilters struct {
	// Disjunctive set of enclave UUIDs for which to return API containers
	// If nil or empty, will match all UUIDs
	EnclaveIDs map[enclave.EnclaveUUID]bool

	// Disjunctive set of statuses that returned API containers must conform to
	// If nil or empty, will match all UUIDs
	Statuses map[container.ContainerStatus]bool
}
