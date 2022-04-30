package user_service_registration

type UserServiceRegistrationFilters struct {
	// Disjunctive IDs to search for
	IDs map[ServiceID]bool
}