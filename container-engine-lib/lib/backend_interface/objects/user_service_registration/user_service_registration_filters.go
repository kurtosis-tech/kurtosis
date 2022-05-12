package user_service_registration

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"

type UserServiceRegistrationFilters struct {
	// Disjunctive set of enclave IDs to find registrations for
	EnclaveIDs map[enclave.EnclaveID]bool

	// Disjunctive IDs to search for
	IDs map[ServiceID]bool
}