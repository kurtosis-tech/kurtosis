package startosis_engine

import (
	"context"
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
func (executor *StartosisExecutor) Execute(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) error {
	executor.mutex.Lock()
	defer executor.mutex.Unlock()
	for index, instruction := range instructions {
		err := instruction.Execute(ctx, executor.environment)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred executing instruction number '%v'", index)
		}
	}
	return nil
}
