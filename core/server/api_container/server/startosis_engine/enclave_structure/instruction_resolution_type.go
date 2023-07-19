package enclave_structure

//go:generate go run github.com/dmarkham/enumer -trimprefix "InstructionResolutionType" -type=InstructionResolutionType
type InstructionResolutionType uint8

const (
	// InstructionEqual means instructions are strictly equal - the new onw can be skipped in favour of the old one
	InstructionEqual InstructionResolutionType = iota

	// InstructionShouldBeRun means the instructions are equivalent, the new one can be re-run in plac of the old one
	InstructionShouldBeRun

	// InstructionUnknown means the two instructions are not equal nor equivalent.
	InstructionUnknown

	// InstructionNotResolvableAbort means one of the compared instructions is fundamentally incompatible with the concept of
	// idempotency. Kurtosis should stop trying to run it in an idempotent way and fall back to the default behaviour
	InstructionNotResolvableAbort
)
