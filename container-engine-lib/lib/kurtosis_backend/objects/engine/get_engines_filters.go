package engine

type GetEnginesFilters struct {
	// Set of engine IDs to find engines for
	IDs map[string]bool

	// Set of EngineStatus that returned engines must conform to
	Statuses map[EngineStatus]bool
}