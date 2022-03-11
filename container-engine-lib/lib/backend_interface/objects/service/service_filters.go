package service

type ServiceFilters struct {
	// Disjunctive set of user service IDs to find user services for
	// If nil or empty, will match all IDs
	IDs map[ServiceID]bool

	// Disjunctive set of user service IDs to find user services for
	// If nil or empty, will match all IDs
	GUIDs map[ServiceGUID]bool
}
