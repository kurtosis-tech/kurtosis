package engine

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"

type EngineFilters struct {
	// Disjunctive set of engine GUIDs to find engines for
	// If nil or empty, will match all IDs
	GUIDs map[EngineGUID]bool

	// Disjunctive set of statuses that returned engines must conform to
	// If nil or empty, will match all statuses
	Statuses map[container_status.ContainerStatus]bool
}