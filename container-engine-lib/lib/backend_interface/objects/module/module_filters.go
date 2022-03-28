package module

type ModuleFilters struct {
	// Disjunctive set of module GUIDs to find modules for
	// If nil or empty, will match all GUIDs
	GUIDs map[ModuleGUID]bool
}
