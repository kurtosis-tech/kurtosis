package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionBaseTest interface {
	// GetId is a unique identifier for this test that will be used in errors when a test fails.
	GetId() string

	// GetInstruction should return the instruction this test is testing
	GetInstruction() (*kurtosis_plan_instruction.KurtosisPlanInstruction, error)

	// GetStarlarkCode should return the serialized starlark code matching the argument dict from GetExpectedArguments
	GetStarlarkCode() (string, error)

	// GetExpectedArguments should return the argument dict matching the Starlark code from GetStarlarkCode
	GetExpectedArguments() (starlark.StringDict, error)
}
