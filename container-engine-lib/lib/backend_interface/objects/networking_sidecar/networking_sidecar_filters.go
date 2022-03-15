package networking_sidecar

type NetworkingSidecarFilters struct {
	// Disjunctive set of networking sidecar GUIDs to find networking sidecar for
	// If nil or empty, will match all IDs
	GUIDs map[NetworkingSidecarGUID]bool
}
