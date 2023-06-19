package kurtosis_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

const (
	Representative    = true
	NotRepresentative = false
)

var (
	NoArgs []starlark.Value
)

type KurtosisInstruction interface {
	// Uuid returns a unique UUID for this instruction.
	// UUID are randomly generated when the instruction is built. Each instruction in the plan will have a different
	// UUID
	Uuid() kurtosis_starlark_framework.InstructionUuid

	GetPositionInOriginalScript() *kurtosis_starlark_framework.KurtosisBuiltinPosition

	GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction

	Execute(ctx context.Context) (*string, error)

	// String is only for easy printing in logs and error messages.
	// Most of the time it will just call GetCanonicalInstruction()
	String() string

	// Hash hashes the instruction.
	// The instruction hash is unique per instruction, provided that the code of the instruction is different. Two
	// instructions with the same code will have the same hash.
	Hash() string

	// ValidateAndUpdateEnvironment validates if the instruction can be applied to an environment, and mutates that
	// environment to reflect how Kurtosis would look like after this instruction is successfully executed.
	ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error
}
