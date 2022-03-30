package networking_sidecar

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"

type NetworkingSidecarFilters struct {
	// Disjunctive set of user service GUIDs to find networking sidecar for
	// If nil or empty, will match all IDs
	GUIDs map[service.ServiceGUID]bool
}
