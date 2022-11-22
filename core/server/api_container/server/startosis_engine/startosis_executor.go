package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

var (
	noInstructionOutput *string
)

type StartosisExecutor struct {
	mutex *sync.Mutex
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor() *StartosisExecutor {
	return &StartosisExecutor{
		mutex: &sync.Mutex{},
	}
}

// Execute executes the list of Kurtosis instructions against the Kurtosis backend
// It serializes each instruction that is executed and returned the list of serialized instruction as a result
// It returns an error if something unexpected happens outside the execution of the script
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction) ([]*kurtosis_core_rpc_api_bindings.KurtosisInstruction, *kurtosis_core_rpc_api_bindings.KurtosisExecutionError) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	var successfullyExecutedInstructions []*kurtosis_core_rpc_api_bindings.KurtosisInstruction
	for index, instruction := range instructions {
		var successfullyExecutedInstruction *kurtosis_core_rpc_api_bindings.KurtosisInstruction
		if !dryRun {
			instructionOutput, err := instruction.Execute(ctx)
			if err != nil {
				executionError := binding_constructors.NewKurtosisExecutionError(stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", index+1, instruction.GetCanonicalInstruction()).Error())
				return successfullyExecutedInstructions, executionError
			}
			successfullyExecutedInstruction = binding_constructors.NewKurtosisInstruction(instruction.GetPositionInOriginalScript().ToAPIType(), instruction.GetCanonicalInstruction(), instructionOutput)
		} else {
			successfullyExecutedInstruction = binding_constructors.NewKurtosisInstruction(instruction.GetPositionInOriginalScript().ToAPIType(), instruction.GetCanonicalInstruction(), noInstructionOutput)
		}
		successfullyExecutedInstructions = append(successfullyExecutedInstructions, successfullyExecutedInstruction)
	}
	return successfullyExecutedInstructions, nil
}
