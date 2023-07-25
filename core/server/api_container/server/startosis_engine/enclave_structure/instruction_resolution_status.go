package enclave_structure

//go:generate go run github.com/dmarkham/enumer -trimprefix "InstructionResolutionStatus" -type=InstructionResolutionStatus
type InstructionResolutionStatus uint8

const (
	// InstructionIsEqual means this instruction is strictly equal to the one it's being compared to.
	InstructionIsEqual InstructionResolutionStatus = iota

	// InstructionIsUpdate means this instruction is the same as the one it's being compared to, but with some
	// changes to its parameters. It should be re-run so that the enclave can be updated.
	InstructionIsUpdate

	// InstructionIsUnknown means the two instructions are completely different.
	InstructionIsUnknown

	// InstructionIsNotResolvableAbort means one of the compared instructions is fundamentally incompatible with the
	// concept of idempotency. Kurtosis should stop trying to run it in an idempotent way and fall back to the default
	// behaviour
	InstructionIsNotResolvableAbort
)
