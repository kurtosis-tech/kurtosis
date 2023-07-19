package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"testing"
)

const (
	noInputParams = "{}"
	noReturnValue = starlark.None
)

type StartosisInterpreterIdempotentTestSuite struct {
	suite.Suite
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
	interpreter            *StartosisInterpreter
}

func (suite *StartosisInterpreterIdempotentTestSuite) SetupTest() {
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	serviceNetwork := service_network.NewMockServiceNetwork(suite.T())
	suite.interpreter = NewStartosisInterpreter(serviceNetwork, suite.packageContentProvider, runtimeValueStore)
}

func TestRunStartosisInterpreterIdempotentTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisInterpreterIdempotentTestSuite))
}

func (suite *StartosisInterpreterIdempotentTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

// Most simple case - replay the same package twice
// Current plan ->     [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that this results in the entire set of instruction being skipped
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_IdenticalPackage() {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	testServiceNetwork := service_network.NewMockServiceNetwork(suite.T())
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := suite.newMockInstruction(instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := suite.newMockInstruction(instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := suite.newMockInstruction(instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), instruction3Str, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction3.IsExecuted())
}

// Add an instruction at the end of a package that was already run
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that the first two instructions are already executed, and the last one is in the new plan marked as not
// executed
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_AppendNewInstruction() {
	script := `
def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := suite.newMockInstruction(instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := suite.newMockInstruction(instruction2Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction2, noReturnValue))

	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.False(suite.T(), scheduledInstruction3.IsExecuted())
}

// Run an instruction inside an enclave that is not empty (other non-related package were run in the past)
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [                                                  `print("instruction3")`]
// Check that the first two instructions are marked as imported from a previous plan, already executed, and the last
// one is in the new plan marked as not executed
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_DisjointInstructionSet() {
	script := `
def run(plan, args):
	plan.print("instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := suite.newMockInstruction(instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := suite.newMockInstruction(instruction2Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction2, noReturnValue))

	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.False(suite.T(), scheduledInstruction3.IsExecuted())
}

// This is a bit of an edge case here, we run only part of the instruction that are identical to what was already run
// Current plan ->     [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`                         ]
// Check that instruction 1 and 2 are part of the new plan, already executed, and instruction3 is ALSO part of the
// plan, but marked as imported from a previous plan
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_ReplacePartOfInstructionWithIdenticalInstruction() {
	script := `
def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := suite.newMockInstruction(instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := suite.newMockInstruction(instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := suite.newMockInstruction(instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction3.IsExecuted())
}

// Submit a package with a update on an instruction located "in the middle" of the package
// Current plan ->     [`print("instruction1")`  `print("instruction2")`      `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// That will result in the concatenation of the two plans because we're not able to properly resolve dependencies yet
// [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`  `print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// The first three are imported from a previous plan, already executed, while the last three are from all new
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_InvalidNewVersionOfThePackage() {
	script := `
def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2_NEW")
	plan.print(msg="instruction3")
`

	instruction1Str := `print(msg="instruction1")`
	instruction1 := suite.newMockInstruction(instruction1Str)
	instruction2Str := `print(msg="instruction2")`
	instruction2 := suite.newMockInstruction(instruction2Str)
	instruction3Str := `print(msg="instruction3")`
	instruction3 := suite.newMockInstruction(instruction3Str)

	currentEnclavePlan := instructions_plan.NewInstructionsPlan()
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction1, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction2, noReturnValue))
	require.Nil(suite.T(), currentEnclavePlan.AddInstruction(instruction3, noReturnValue))

	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		currentEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 6, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction1.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction1.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), instruction2Str, scheduledInstruction2.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction2.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), instruction3Str, scheduledInstruction3.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction3.IsImportedFromCurrentEnclavePlan())
	require.True(suite.T(), scheduledInstruction3.IsExecuted())

	scheduledInstruction4 := instructionSequence[3]
	require.Equal(suite.T(), instruction1Str, scheduledInstruction4.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction4.IsImportedFromCurrentEnclavePlan())
	require.False(suite.T(), scheduledInstruction4.IsExecuted())

	scheduledInstruction5 := instructionSequence[4]
	require.Equal(suite.T(), `print(msg="instruction2_NEW")`, scheduledInstruction5.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction5.IsImportedFromCurrentEnclavePlan())
	require.False(suite.T(), scheduledInstruction5.IsExecuted())

	scheduledInstruction6 := instructionSequence[5]
	require.Equal(suite.T(), instruction3Str, scheduledInstruction6.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction6.IsImportedFromCurrentEnclavePlan())
	require.False(suite.T(), scheduledInstruction6.IsExecuted())
}

func (suite *StartosisInterpreterIdempotentTestSuite) newMockInstruction(instructionName string) kurtosis_instruction.KurtosisInstruction {
	mockInstruction := mock_instruction.NewMockKurtosisInstruction(suite.T())
	mockInstruction.EXPECT().String().Maybe().Return(instructionName)
	return mockInstruction
}
