package repl

type ReplFilters struct {
	// Disjunctive set of repl IDs to find repls for
	// If nil or empty, will match all IDs
	GUIDs map[ReplGUID]bool
}
