package enclave_statuses

type EnclaveStatus string

const (
	// One or more containers in the enclave is still running
	Running EnclaveStatus = "RUNNING"

	// All containers in the enclave are stopped
	Stopped EnclaveStatus = "STOPPED"
)
