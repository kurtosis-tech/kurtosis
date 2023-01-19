package networking_sidecar

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type NetworkingSidecarFilters struct {
	// Disjunctive set of enclave UUIDs for which to return networking sidecars
	// If nil or empty, will match all enclave UUIDs
	EnclaveUUIDs map[enclave.EnclaveUUID]bool

	// Disjunctive set of user service UUIDs to find networking sidecars for
	// If nil or empty, will match all UUIDs
	UserServiceUUIDs map[service.ServiceUUID]bool

	// Disjunctive set of statuses that returned networking sidecars must conform to
	// If nil or empty, will match all statuses
	Statuses map[container_status.ContainerStatus]bool
}
