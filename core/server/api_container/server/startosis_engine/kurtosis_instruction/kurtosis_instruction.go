package kurtosis_instruction

type RawKurtosisInstruction struct {
	instruction string
}

type KurtosisInstructionOutput struct {
	Output string
}

type KurtosisInstruction interface {
	GetRawInstruction() *RawKurtosisInstruction

	Execute() (*KurtosisInstructionOutput, error)

	String() string
}
