package api_container

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
)

type APIContainerFilters struct {
	// Disjunctive set of enclave IDs for which to return API containers
	// If nil or empty, will match all IDs
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive set of statuses that returned API containers must conform to
	// If nil or empty, will match all IDs
	Statuses map[container_status.ContainerStatus]bool
}