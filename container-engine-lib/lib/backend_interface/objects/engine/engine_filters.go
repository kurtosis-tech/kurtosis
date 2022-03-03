package engine

type EngineFilters struct {
	// Disjunctive set of engine IDs to find engines for
	// If nil or empty, will match all IDs
	IDs map[string]bool

	// Disjunctive set of EngineStatus that returned engines must conform to
	// If nil or empty, will match all IDs
	Statuses map[EngineStatus]bool
}