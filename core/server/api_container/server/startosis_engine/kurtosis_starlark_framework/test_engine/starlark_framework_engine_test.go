package test_engine

import (
	"context"
	"fmt"
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

const (
	resultStarlarkVar = "result"
)

func TestAllRegisteredBuiltins(t *testing.T) {
	testsAllKurtosisPlanInstructions(t, newAddServiceTestCase(t))
	testsAllKurtosisPlanInstructions(t, newAddServicesTestCase(t))
	testsAllKurtosisPlanInstructions(t, newAssertTestCase(t))
	testsAllKurtosisPlanInstructions(t, newExecTestCase(t))
	testsAllKurtosisPlanInstructions(t, newSetConnectionTestCase(t))
	testsAllKurtosisPlanInstructions(t, newSetConnectionDefaultTestCase(t))
	testsAllKurtosisPlanInstructions(t, newRenderTemplateTestCase1(t))
	testsAllKurtosisPlanInstructions(t, newRenderTemplateTestCase2(t))
}

func testsAllKurtosisPlanInstructions(t *testing.T, builtin KurtosisPlanInstructionBaseTest) {
	testId := builtin.GetId()
	var instructionQueue []kurtosis_instruction.KurtosisInstruction
	thread := shared_helpers.NewStarlarkThread("framework-testing-engine")

	predeclared := getBasePredeclaredDict()
	// Add the KurtosisPlanInstruction that is being tested
	instructionFromBuiltin := builtin.GetInstruction()
	instructionWrapper := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(instructionFromBuiltin, &instructionQueue)
	predeclared[instructionWrapper.GetName()] = starlark.NewBuiltin(add_service.AddServiceBuiltinName, instructionWrapper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), predeclared)
	require.Nil(t, err, "Error interpreting Starlark code for instruction '%s'", testId)
	interpretationResult := extractResultValue(t, globals)

	instruction, ok := instructionQueue[0].(*kurtosis_plan_instruction.KurtosisPlanInstructionInternal)
	require.True(t, ok, "Builtin expected to be a KurtosisPlanInstructionInternal, but was '%s'", reflect.TypeOf(instruction))

	// execute the instruction and run custom builtin assertions
	executionResult, err := instruction.Execute(context.Background())
	require.Nil(t, err, "Builtin execution threw an error: \n%v", err)
	builtin.Assert(interpretationResult, executionResult)

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

func codeToExecute(builtinStarlarkCode string) string {
	return fmt.Sprintf("%s = %s", resultStarlarkVar, builtinStarlarkCode)
}

func extractResultValue(t *testing.T, globals starlark.StringDict) starlark.Value {
	value, found := globals[resultStarlarkVar]
	require.True(t, found, "Result variable could not be found in dictionary of global variables")
	return value
}
