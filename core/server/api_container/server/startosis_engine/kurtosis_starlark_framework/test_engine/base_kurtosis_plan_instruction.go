package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionBaseTest interface {
	// GetInstruction should return the instruction this test is testing
	GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction

	// GetStarlarkCode should return the serialized starlark code matching the argument dict from GetExpectedArguments
	GetStarlarkCode() string

	// GetStarlarkCodeForAssertion sometimes, for instance when testing positional args, we need a starlark
	//code for execution (which will be provided by GetStarlarkCode) and another version for assertion
	//this method provide the second one,
	//this can be used only if having different scripts (one for execution and another for assertion) is needed, otherwise this should return an empty string
	GetStarlarkCodeForAssertion() string

	// Assert is called after the Starlark code returned by GetStarlarkCode has been sent to the interpreter
	// The interpretationResult argument corresponds to the object returns by the interpreter
	// The executionResult argument corresponds to the execution result string
	Assert(interpretationResult starlark.Value, executionResult *string)
}
