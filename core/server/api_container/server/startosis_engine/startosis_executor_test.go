package startosis_engine

import (
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	executeSuccessfully = true
	throwOnExecute      = false

	doDryRun       = true
	executeForReal = false

	noScriptOutputObject = ""
	noParallelism        = 1
)

var (
	dummyPosition               = kurtosis_instruction.NewInstructionPosition(12, 1, "dummyFile")
	noInstructionArgsForTesting []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg
)

func TestExecuteKurtosisInstructions_ExecuteForReal_Success(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
	}

	_, serializedInstruction, err := executeSynchronously(t, executor, executeForReal, instructions)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting),
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting),
	}
	require.Equal(t, expectedSerializedInstructions, serializedInstruction)
}

func TestExecuteKurtosisInstructions_ExecuteForReal_FailureHalfWay(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2", throwOnExecute)
	instruction3 := createMockInstruction(t, "instruction3", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
		instruction3,
	}

	_, serializedInstruction, executionError := executeSynchronously(t, executor, executeForReal, instructions)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 1)
	instruction2.AssertNumberOfCalls(t, "String", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)
	// nothing called for instruction 3 because instruction 2 threw an error
	instruction3.AssertNumberOfCalls(t, "GetCanonicalInstruction", 0)
	instruction3.AssertNumberOfCalls(t, "Execute", 0)

	expectedErrorMsgPrefix := `An error occurred executing instruction (number 2): 
instruction2()
 --- at`
	expectedLowLevelErrorMessage := "expected error for test"
	require.NotNil(t, executionError)
	require.Contains(t, executionError.GetErrorMessage(), expectedErrorMsgPrefix)
	require.Contains(t, executionError.GetErrorMessage(), expectedLowLevelErrorMessage)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		// only instruction 1 and 2 because it failed at instruction 2
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting),
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting),
	}
	require.Equal(t, expectedSerializedInstructions, serializedInstruction)
}

func TestExecuteKurtosisInstructions_DoDryRun(t *testing.T) {
	executor := NewStartosisExecutor()

	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2", executeSuccessfully)
	instructions := []kurtosis_instruction.KurtosisInstruction{
		instruction1,
		instruction2,
	}

	_, serializedInstruction, err := executeSynchronously(t, executor, doDryRun, instructions)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	// both execute never called because dry run = true
	instruction1.AssertNumberOfCalls(t, "Execute", 0)
	instruction2.AssertNumberOfCalls(t, "Execute", 0)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting),
		binding_constructors.NewStarlarkInstruction(dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting),
	}
	require.Equal(t, serializedInstruction, expectedSerializedInstructions)
}

func createMockInstruction(t *testing.T, instructionName string, executeSuccessfully bool) *mock_instruction.MockKurtosisInstruction {
	instruction := mock_instruction.NewMockKurtosisInstruction(t)

	stringifiedInstruction := instructionName + "()"
	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		dummyPosition.ToAPIType(), instructionName, stringifiedInstruction, noInstructionArgsForTesting)
	instruction.EXPECT().GetCanonicalInstruction().Maybe().Return(canonicalInstruction)
	instruction.EXPECT().GetPositionInOriginalScript().Maybe().Return(dummyPosition)
	instruction.EXPECT().String().Maybe().Return(stringifiedInstruction)

	if executeSuccessfully {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, nil)
	} else {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, errors.New("expected error for test"))
	}

	return instruction
}

func executeSynchronously(t *testing.T, executor *StartosisExecutor, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction) (string, []*kurtosis_core_rpc_api_bindings.StarlarkInstruction, *kurtosis_core_rpc_api_bindings.StarlarkExecutionError) {
	scriptOutput := strings.Builder{}
	var serializedInstructions []*kurtosis_core_rpc_api_bindings.StarlarkInstruction
	executionResponseLines := executor.Execute(context.Background(), dryRun, noParallelism, instructions, noScriptOutputObject)
	for executionResponseLine := range executionResponseLines {
		if executionResponseLine.GetError() != nil {
			return scriptOutput.String(), serializedInstructions, executionResponseLine.GetError().GetExecutionError()
		}
		if executionResponseLine.GetInstruction() != nil {
			executedKurtosisInstruction := executionResponseLine.GetInstruction()
			serializedInstructions = append(serializedInstructions, executedKurtosisInstruction)
		}
		if executionResponseLine.GetInstructionResult() != nil {
			if _, err := scriptOutput.WriteString(executionResponseLine.GetInstructionResult().GetSerializedInstructionResult()); err != nil {
				require.Nil(t, err)
			}
		}
	}
	return scriptOutput.String(), serializedInstructions, nil
}
