package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
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

// Execute executes the list of Kurtosis instructions against the Kurtosis backend
// It serializes each instruction that is executed and returned the list of serialized instruction as a result
// It returns an error if something unexpected happens outside the execution of the script
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction, outputBuffer *strings.Builder) ([]*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction, error) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	var serializedSuccessfullyExecutedInstructions []*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction
	for index, instruction := range instructions {
		if !dryRun {
			instructionOutput, err := instruction.Execute(ctx)
			if err != nil {
				return serializedSuccessfullyExecutedInstructions, stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", index+1, instruction.GetCanonicalInstruction())
			}
			if instructionOutput != nil {
				outputBuffer.WriteString(*instructionOutput)
			}
		}
		serializedSuccessfullyExecutedInstructions = append(serializedSuccessfullyExecutedInstructions,
			binding_constructors.NewSerializedKurtosisInstruction(instruction.GetCanonicalInstruction()))
	}
	return serializedSuccessfullyExecutedInstructions, nil
}
