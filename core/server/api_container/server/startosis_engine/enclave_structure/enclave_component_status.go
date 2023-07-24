package enclave_structure

//go:generate go run github.com/dmarkham/enumer -trimprefix "EnclaveComponentStatus" -type=EnclaveComponentStatus
type EnclaveComponentStatus uint8

const (
	// ComponentIsNew means the component was first created during this run
	ComponentIsNew EnclaveComponentStatus = iota

	// ComponentWasLeftIntact means the component was present prior to the beginning of this run and has been left
	// intact
	ComponentWasLeftIntact

	// ComponentIsUpdated means the component was present prior to the beginning of this run but has been updated
	// during this run
	ComponentIsUpdated
)
