package enclave

type EnclaveFilters struct {
	// Disjunctive set of enclave IDs to operate on
	// If nil or empty, will match all IDs
	IDs map[EnclaveID]bool

	// Disjunctive set of enclave statuses to operate on
	// If nil or empty, will match all IDs
	Statuses map[EnclaveStatus]bool
}
