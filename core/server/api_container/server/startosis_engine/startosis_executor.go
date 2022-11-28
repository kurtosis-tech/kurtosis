package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
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

// Execute executes the list of Kurtosis instructions _asynchronously_ against the Kurtosis backend
// Consumers of this method should read the response lines channel and return as soon as one it is closed
//
// The channel of KurtosisExecutionResponseLine can contain three kinds of line:
// - A regular KurtosisInstruction that was successfully executed
// - A KurtosisExecutionError if the execution failed
// - A ProgressInfo to update the current "state" of the execution
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction) <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine {
	executor.mutex.Lock()
	kurtosisExecutionResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine)

	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(kurtosisExecutionResponseLineStream)
		}()

		totalNumberOfInstructions := uint32(len(instructions))
		for index, instruction := range instructions {
			instructionNumber := uint32(index + 1)
			kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromProgressInfo(
				"Execution in progress", instructionNumber, totalNumberOfInstructions)
			if !dryRun {
				instructionOutput, err := instruction.Execute(ctx)
				if err != nil {
					propagatedError := stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", instructionNumber, instruction.String())
					serializedError := binding_constructors.NewKurtosisExecutionError(propagatedError.Error())
					kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromExecutionError(serializedError)
					return
				}
				instructionWithResult := binding_constructors.AddResultToKurtosisInstruction(instruction.GetCanonicalInstruction(), instructionOutput)
				kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromInstruction(instructionWithResult)
			} else {
				instructionWithoutResult := instruction.GetCanonicalInstruction()
				kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromInstruction(instructionWithoutResult)
			}
		}
	}()
	return kurtosisExecutionResponseLineStream
}
