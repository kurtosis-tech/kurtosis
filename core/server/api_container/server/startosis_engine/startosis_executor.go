package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	progressMsg      = "Execution in progress"
	ParallelismParam = "PARALLELISM"
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
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, parallelism int, instructions []kurtosis_instruction.KurtosisInstruction, serializedScriptOutput string) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	executor.mutex.Lock()
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	ctxWithParallelism := context.WithValue(ctx, ParallelismParam, parallelism)
	go func() {
		defer func() {
			executor.mutex.Unlock()
			close(starlarkRunResponseLineStream)
		}()

		totalNumberOfInstructions := uint32(len(instructions))
		for index, instruction := range instructions {
			instructionNumber := uint32(index + 1)
			progress := binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
				progressMsg, instructionNumber, totalNumberOfInstructions)
			starlarkRunResponseLineStream <- progress

			canonicalInstruction := binding_constructors.NewStarlarkRunResponseLineFromInstruction(instruction.GetCanonicalInstruction())
			starlarkRunResponseLineStream <- canonicalInstruction

			if !dryRun {
				instructionOutput, err := instruction.Execute(ctxWithParallelism)
				if err != nil {
					propagatedError := stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", instructionNumber, instruction.String())
					serializedError := binding_constructors.NewStarlarkExecutionError(propagatedError.Error())
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromExecutionError(serializedError)
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
					return
				}
				if instructionOutput != nil {
					starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(*instructionOutput)
				}
			}
		}

		// TODO(gb): we should run magic string replacement on the output
		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(serializedScriptOutput)
	}()
	return starlarkRunResponseLineStream
}
