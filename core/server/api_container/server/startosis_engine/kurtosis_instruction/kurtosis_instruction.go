package kurtosis_instruction

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"

type RawKurtosisInstruction struct {
	instruction string
}

type KurtosisInstructionOutput struct {
	Output string
}

type KurtosisInstruction interface {
	GetRawInstruction() *RawKurtosisInstruction

	Execute(backend backend_interface.KurtosisBackend) (*KurtosisInstructionOutput, error)

	String() string
}
