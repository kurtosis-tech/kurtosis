package enclave_states

type EnclaveState string

const (
	// One or more containers in the enclave is still running
	Running EnclaveState = "RUNNING"

	// All containers in the enclave are stopped
	Stopped EnclaveState = "STOPPED"
)
