package module

type ModuleFilters struct {
	// Disjunctive set of module IDs to find modules for
	// If nil or empty, will match all IDs
	IDs map[string]bool
}
