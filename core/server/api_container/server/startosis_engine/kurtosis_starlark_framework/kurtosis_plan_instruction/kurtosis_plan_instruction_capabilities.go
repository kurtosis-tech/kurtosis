package kurtosis_plan_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionCapabilities interface {
	Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError)

	Validate(arguments *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError

	Execute(ctx context.Context, arguments *builtin_argument.ArgumentValuesSet) (string, error)

	// IsEquivalentTo returns whether two instructions are equivalent.
	// See KurtosisInstruction for the definition of equivalent instructions
	TryResolveWith(areEquals bool, other KurtosisPlanInstructionCapabilities, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionType
}
