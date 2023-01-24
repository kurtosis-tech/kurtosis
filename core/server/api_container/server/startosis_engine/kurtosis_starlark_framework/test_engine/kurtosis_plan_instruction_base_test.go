package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionBaseTest interface {
	// GetId is a unique identifier for this test that will be used in errors when a test fails.
	GetId() string

	// GetInstruction should return the instruction this test is testing
	GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction

	// GetStarlarkCode should return the serialized starlark code matching the argument dict from GetExpectedArguments
	GetStarlarkCode() string

	// Assert is called after the Starlark code returned by GetStarlarkCode has been sent to the interpreter
	// The interpretationResult argument corresponds to the object returns by the interpreter
	// The executionResult argument corresponds to the execution result string
	Assert(interpretationResult starlark.Value, executionResult *string)
}
