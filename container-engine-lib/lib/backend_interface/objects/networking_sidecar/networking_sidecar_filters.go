package networking_sidecar

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"

type NetworkingSidecarFilters struct {
	EnclaveId enclave.EnclaveID

	// Disjunctive set of networking sidecar GUIDs to find networking sidecar for
	// If nil or empty, will match all IDs
	GUIDs map[NetworkingSidecarGUID]bool
}
