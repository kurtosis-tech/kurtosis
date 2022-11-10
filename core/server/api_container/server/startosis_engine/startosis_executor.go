package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

type StartosisExecutor struct {
	mutex       *sync.Mutex
	environment *startosis_executor.ExecutionEnvironment
}

type ExecutionError struct {
	Error string
}

func NewStartosisExecutor() *StartosisExecutor {
	return &StartosisExecutor{
		mutex:       &sync.Mutex{},
		environment: startosis_executor.NewExecutionEnvironment(),
	}
}

// Execute executes the list of Kurtosis instructions against the Kurtosis backend
// It returns a potential execution error if something went wrong.
// It returns an error if something unexpected happens outside the execution of the script
func (executor *StartosisExecutor) Execute(ctx context.Context, dryRun bool, instructions []kurtosis_instruction.KurtosisInstruction) ([]*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction, error) {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()

	var serializedSuccessfullyExecutedInstructions []*kurtosis_core_rpc_api_bindings.SerializedKurtosisInstruction
	for index, instruction := range instructions {
		if !dryRun {
			if err := instruction.Execute(ctx, executor.environment); err != nil {
				return serializedSuccessfullyExecutedInstructions, stacktrace.Propagate(err, "An error occurred executing instruction (number %d): \n%v", index+1, instruction.GetCanonicalInstruction())
			}
		}
		serializedSuccessfullyExecutedInstructions = append(serializedSuccessfullyExecutedInstructions,
			binding_constructors.NewSerializedKurtosisInstruction(instruction.GetCanonicalInstruction()))
	}
	return serializedSuccessfullyExecutedInstructions, nil
}
