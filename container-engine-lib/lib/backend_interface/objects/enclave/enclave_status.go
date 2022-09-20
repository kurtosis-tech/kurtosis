package enclave

//go:generate go run github.com/dmarkham/enumer -trimprefix=EnclaveStatus_ -transform=snake-upper -type=EnclaveStatus
type EnclaveStatus int
const (
	EnclaveStatus_Empty EnclaveStatus = iota   // No containers exist inside the enclave
	EnclaveStatus_Running	// The enclave has containers, and at least one container is running
	EnclaveStatus_Stopped	// The enclave has containers, but they're all stopped
)
