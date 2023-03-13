package enclave

type EnclaveFilters struct {
	// Disjunctive set of enclave UUIDs to operate on
	// If nil or empty, will match all UUIDs
	UUIDs map[EnclaveUUID]bool

	// Disjunctive set of enclave statuses to operate on
	// If nil or empty, will match all UUIDs
	Statuses map[EnclaveStatus]bool
}
