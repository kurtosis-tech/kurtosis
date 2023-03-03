package test_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"testing"
)

const (
	resultStarlarkVar = "result"
)

func TestAllRegisteredBuiltins(t *testing.T) {
	testKurtosisPlanInstruction(t, newAddServiceTestCase(t))
	testKurtosisPlanInstruction(t, newAddServicesTestCase(t))
	testKurtosisPlanInstruction(t, newAssertTestCase(t))
	testKurtosisPlanInstruction(t, newExecTestCase1(t))
	testKurtosisPlanInstruction(t, newExecTestCase2(t))
	testKurtosisPlanInstruction(t, newSetConnectionTestCase(t))
	testKurtosisPlanInstruction(t, newSetConnectionDefaultTestCase(t))
	testKurtosisPlanInstruction(t, newRemoveConnectionTestCase(t))
	testKurtosisPlanInstruction(t, newRemoveServiceTestCase(t))
	testKurtosisPlanInstruction(t, newRenderSingleTemplateTestCase(t))
	testKurtosisPlanInstruction(t, newRenderMultipleTemplatesTestCase(t))
	testKurtosisPlanInstruction(t, newRequestTestCase1(t))
	testKurtosisPlanInstruction(t, newRequestTestCase2(t))
	testKurtosisPlanInstruction(t, newStoreServiceFilesTestCase(t))
	testKurtosisPlanInstruction(t, newStoreServiceFilesWithoutNameTestCase(t))
	testKurtosisPlanInstruction(t, newUpdateServiceTestCase(t))
	testKurtosisPlanInstruction(t, newUploadFilesTestCase(t))
	testKurtosisPlanInstruction(t, newUploadFilesWithoutNameTestCase(t))
	testKurtosisPlanInstruction(t, newWaitTestCase1(t))
	testKurtosisPlanInstruction(t, newWaitTestCase2(t))

	testKurtosisHelper(t, newReadFileTestCase(t))
	testKurtosisHelper(t, newImportModuleTestCase(t))

	testKurtosisTypeConstructor(t, newUpdateServiceConfigTestCase(t))
}

func testKurtosisPlanInstruction(t *testing.T, builtin KurtosisPlanInstructionBaseTest) {
	testId := builtin.GetId()
	var instructionQueue []kurtosis_instruction.KurtosisInstruction
	thread := newStarlarkThread("framework-testing-engine")

	predeclared := getBasePredeclaredDict()
	// Add the KurtosisPlanInstruction that is being tested
	instructionFromBuiltin := builtin.GetInstruction()
	instructionWrapper := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(instructionFromBuiltin, &instructionQueue)
	predeclared[instructionWrapper.GetName()] = starlark.NewBuiltin(instructionWrapper.GetName(), instructionWrapper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), predeclared)
	require.Nil(t, err, "Error interpreting Starlark code for instruction '%s'", testId)
	interpretationResult := extractResultValue(t, globals)

	require.Len(t, instructionQueue, 1)
	instructionToExecute := instructionQueue[0]

	// execute the instruction and run custom builtin assertions
	executionResult, err := instructionToExecute.Execute(context.WithValue(context.Background(), "PARALLELISM", 1))
	require.Nil(t, err, "Builtin execution threw an error: \n%v", err)
	builtin.Assert(interpretationResult, executionResult)

	// check serializing the obtained instruction falls back to the initial one
	serializedInstruction := instructionToExecute.String()

	starlarkCodeForAssertion := builtin.GetStarlarkCodeForAssertion()
	if starlarkCodeForAssertion == "" {
		starlarkCodeForAssertion = starlarkCode
	}

	require.Equal(t, starlarkCodeForAssertion, serializedInstruction)
}

func testKurtosisHelper(t *testing.T, builtin KurtosisHelperBaseTest) {
	testId := builtin.GetId()
	thread := newStarlarkThread("framework-testing-engine")

	predeclared := getBasePredeclaredDict()
	// Add the KurtosisPlanInstruction that is being tested
	helper := builtin.GetHelper()
	predeclared[helper.GetName()] = starlark.NewBuiltin(helper.GetName(), helper.CreateBuiltin())

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), predeclared)
	require.Nil(t, err, "Error interpreting Starlark code for builtin '%s'", testId)
	result := extractResultValue(t, globals)

	builtin.Assert(result)
}

func testKurtosisTypeConstructor(t *testing.T, builtin KurtosisTypeConstructorBaseTest) {
	testId := builtin.GetId()
	thread := newStarlarkThread("framework-testing-engine")

	predeclared := getBasePredeclaredDict()

	starlarkCode := builtin.GetStarlarkCode()
	globals, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, codeToExecute(starlarkCode), predeclared)
	require.Nil(t, err, "Error interpreting Starlark code for builtin '%s'", testId)
	result := extractResultValue(t, globals)

	builtin.Assert(result)

	serializedType := result.String()
	require.Equal(t, starlarkCode, serializedType)
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

func newStarlarkThread(name string) *starlark.Thread {
	return &starlark.Thread{
		Name:       name,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
}
