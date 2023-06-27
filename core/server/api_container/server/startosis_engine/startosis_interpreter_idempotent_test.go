package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	noInputParams = "{}"
	noReturnValue = starlark.None
)

// Most simple case - replay the same package twice
// Current plan ->     [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that this results in the entire set of instruction being skipped
func TestInterpretAndOptimize_IdenticalPackage(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := newMockInstruction(t, instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := newMockInstruction(t, instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := newMockInstruction(t, instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(t, interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	require.Equal(t, 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(t, instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(t, scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(t, instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(t, scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(t, instruction3Str, scheduledInstruction3.GetInstruction().String())
	require.False(t, scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction3.IsExecuted())
}

// Add an instruction at the end of a package that was already run
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that the first two instructions are a;ready executed, and the last one is in the new plan marked as not
// executed
func TestInterpretAndOptimize_AppendNewInstruction(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := newMockInstruction(t, instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := newMockInstruction(t, instruction2Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction2, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(t, interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	require.Equal(t, 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(t, instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(t, scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(t, instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(t, scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(t, `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(t, scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.False(t, scheduledInstruction3.IsExecuted())
}

// Run an instruction inside an enclave that is not empty (other non-related package were run in the past)
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [                                                  `print("instruction3")`]
// Check that the first two instructions are marked as imported from a previous plan, already executed, and the last
// one is in the new plan marked as not executed
func TestInterpretAndOptimize_DisjointInstructionSet(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := newMockInstruction(t, instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := newMockInstruction(t, instruction2Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction2, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(t, interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	require.Equal(t, 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(t, instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.True(t, scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(t, instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.True(t, scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(t, `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(t, scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.False(t, scheduledInstruction3.IsExecuted())
}

// This is a bit of an edge case here, we run only part of the instruction that are identical to what was already run
// Current plan ->     [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`                         ]
// Check that instruction 1 and 2 are part of the new plan, already executed, and instruction3 is ALSO part of the
// plan, but marked as imported from a previous plan
func TestInterpretAndOptimize_ReplacePartOfInstructionWithIdenticalInstruction(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := newMockInstruction(t, instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := newMockInstruction(t, instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := newMockInstruction(t, instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(t, interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	require.Equal(t, 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(t, instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(t, scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(t, instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(t, scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(t, `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.True(t, scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction3.IsExecuted())
}

// Submit a package with a update on an instruction located "in the middle" of the package
// Current plan ->     [`print("instruction1")`  `print("instruction2")`      `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// That will result in the concatenation of the two plans because we're not able to properly resolve dependencies yet
// [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`  `print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// The first three are imported from a previous plan, already executed, while the last three are from all new
func TestInterpretAndOptimize_InvalidNewVersionOfThePackage(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2_NEW")
	plan.print(msg="instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := newMockInstruction(t, instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := newMockInstruction(t, instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := newMockInstruction(t, instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(t, currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(t, interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	require.Equal(t, 6, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(t, instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.True(t, scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(t, instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.True(t, scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(t, instruction3Str, scheduledInstruction3.GetInstruction().String())
	require.True(t, scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(t, scheduledInstruction3.IsExecuted())

	scheduledInstruction4 := instructionSequence[3]
	require.Equal(t, instruction1Str, scheduledInstruction4.GetInstruction().String())
	require.False(t, scheduledInstruction4.IsImportedFromCurrentEnclavePlan())
	require.False(t, scheduledInstruction4.IsExecuted())

	scheduledInstruction5 := instructionSequence[4]
	require.Equal(t, `print(msg="instruction2_NEW")`, scheduledInstruction5.GetInstruction().String())
	require.False(t, scheduledInstruction5.IsImportedFromCurrentEnclavePlan())
	require.False(t, scheduledInstruction5.IsExecuted())

	scheduledInstruction6 := instructionSequence[5]
	require.Equal(t, instruction3Str, scheduledInstruction6.GetInstruction().String())
	require.False(t, scheduledInstruction6.IsImportedFromCurrentEnclavePlan())
	require.False(t, scheduledInstruction6.IsExecuted())
}

func newMockInstruction(t *testing.T, instructionName string) kurtosis_instruction.KurtosisInstruction {
	mockInstruction := mock_instruction.NewMockKurtosisInstruction(t)
	mockInstruction.EXPECT().String().Maybe().Return(instructionName)
	return mockInstruction
}
