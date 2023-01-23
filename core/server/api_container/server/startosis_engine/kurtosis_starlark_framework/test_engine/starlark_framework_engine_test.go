package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"reflect"
	"testing"
)

var registeredBuiltin = []KurtosisPlanInstructionBaseTest{
	renderTemplateTestCase1{},
	renderTemplateTestCase2{},
}

func TestAllRegisteredBuiltins(t *testing.T) {
	for _, builtin := range registeredBuiltin {
		testsAllKurtosisPlanInstructions(t, builtin)
	}
}

func testsAllKurtosisPlanInstructions(t *testing.T, builtin KurtosisPlanInstructionBaseTest) {
	testId := builtin.GetId()
	var instructionQueue []kurtosis_instruction.KurtosisInstruction
	thread := shared_helpers.NewStarlarkThread("framework-testing-engine")

	predeclared := getBasePredeclaredDict()
	// Add the KurtosisPlanInstruction that is being tested
	instructionFromBuiltin, err := builtin.GetInstruction()
	require.Nilf(t, err, "Error retrieving instruction from builtin '%s'", testId)
	instructionWrapper := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(instructionFromBuiltin, &instructionQueue)
	predeclared[instructionWrapper.GetName()] = starlark.NewBuiltin(add_service.AddServiceBuiltinName, instructionWrapper.CreateBuiltin())

	starlarkCode, err := builtin.GetStarlarkCode()
	require.Nilf(t, err, "Error retrieving Starlark code from builtin '%s'", testId)
	_, err = starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkCode, predeclared)
	require.Nil(t, err, "Error interpreting Starlark code for instruction '%s'", testId)

	instruction, ok := instructionQueue[0].(*kurtosis_plan_instruction.KurtosisPlanInstructionInternal)
	require.True(t, ok, "Builtin expected to be a KurtosisPlanInstructionInternal, but was '%s'", reflect.TypeOf(instruction))

	// check arguments matches the expected
	expectedArguments, err := builtin.GetExpectedArguments()
	require.Nilf(t, err, "Error retrieving expected argument from builtin '%s'", testId)
	for _, expectedArgumentName := range expectedArguments.Keys() {
		expectedArgumentValue, found := expectedArguments[expectedArgumentName]
		require.True(t, found, "Unexpected error happened iterating over the argument map - key '%s' was supposed to exist in the Starlark dictionary '%v'", expectedArgumentName, expectedArguments)

		var actualValue starlark.Value
		extractionErr := instruction.GetArguments().ExtractArgumentValue(expectedArgumentName, &actualValue)
		require.Nil(t, extractionErr, "Argument '%s' was supposed to exist in the actual arguments values set, but was not. Actual arguments were: '%s'", expectedArgumentName, instruction.GetArguments().String())

		require.Equal(t, expectedArgumentValue, actualValue, "Argument value did not match the expected for test '%s'", testId)
	}

	// check serializing the obtained instruction falls back to the initial one
	serializedInstruction := instruction.String()
	require.Equal(t, starlarkCode, serializedInstruction)
}

func getBasePredeclaredDict() starlark.StringDict {
	// TODO: refactor this with the one we have in the interpreter
	predeclared := starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		time.Module.Name:                  time.Module,

		// Kurtosis pre-built module containing Kurtosis constant types
		builtins.KurtosisModuleName: builtins.KurtosisModule(),
	}
	// Add all Kurtosis types
	for _, kurtosisTypeConstructors := range startosis_engine.KurtosisTypeConstructors() {
		predeclared[kurtosisTypeConstructors.Name()] = kurtosisTypeConstructors
	}
	return predeclared
}
