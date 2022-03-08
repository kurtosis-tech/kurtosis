package service

type ServiceFilters struct {
	// Disjunctive set of user service IDs to find user services for
	// If nil or empty, will match all IDs
	IDs map[string]bool
}
