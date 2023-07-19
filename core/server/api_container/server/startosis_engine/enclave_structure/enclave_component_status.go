package enclave_structure

//go:generate go run github.com/dmarkham/enumer -trimprefix "EnclaveComponentStatus" -type=EnclaveComponentStatus
type EnclaveComponentStatus uint8

const (
	// ComponentNew means the component was first created during this run
	ComponentNew EnclaveComponentStatus = iota

	// ComponentLeftIntact means the component was present prior to the beginning of this run and has been left intact
	ComponentLeftIntact

	// ComponentUpdated means the component was present prior to the beginning of this run but has been updated during this run
	ComponentUpdated
)
