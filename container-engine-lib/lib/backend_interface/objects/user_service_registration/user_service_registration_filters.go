package user_service_registration

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"

type UserServiceRegistrationFilters struct {
	// Disjunctive set of registration GUIDs to search for
	GUIDs map[UserServiceRegistrationGUID]bool

	// Disjunctive set of enclave IDs to find registrations for
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive IDs to search for
	ServiceIDs map[ServiceID]bool
}