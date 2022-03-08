package enclave

type EnclaveFilters struct {
	// Disjunctive set of enclave IDs to operate on
	IDs map[string]bool

	// Disjunctive set of enclave statuses to operate on
	Statuses map[EnclaveStatus]bool
}
